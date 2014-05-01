package resolver

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelslice"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"
	"github.com/mikobautista/tartantrades/set"

	_ "github.com/ziutek/mymysql/godrv"
)

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG = logging.NewLogger(true)
)

type user struct {
	id       uint
	username string
	password string
	token    string
}

type Token struct {
	Username   string
	Experation time.Time
}

type ResolverServer struct {
	db              *sql.DB
	connectionMap   channelmap.ChannelMap
	apiSevers       channelslice.ChannelSlice
	tradeServers    channelmap.ChannelMap
	tcpPort         int
	httpPort        int
	tableName       string
	dbUser          string
	dbPw            string
	sessionDuration int
	checkExpires    bool
	servers         *set.IntSet
}

func NewResolverServer(
	httpPort int,
	tcpPort int,
	tableName string,
	dbUser string,
	dbPw string,
	sessionDuration int,
	checkExpires bool) *ResolverServer {

	return &ResolverServer{
		connectionMap:   channelmap.NewChannelMap(),
		apiSevers:       channelslice.NewChannelSlice(),
		tradeServers:    channelmap.NewChannelMap(),
		tcpPort:         tcpPort,
		httpPort:        httpPort,
		tableName:       tableName,
		dbUser:          dbUser,
		dbPw:            dbPw,
		sessionDuration: sessionDuration,
		checkExpires:    checkExpires,
		servers:         set.NewUint32Set(),
	}
}

func (rs *ResolverServer) Start() {
	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", rs.tableName, rs.dbUser, rs.dbPw))
	rs.db = db
	LOG.CheckForError(err, true)
	defer db.Close()

	m := make(map[string]shared.HandlerType)
	m["/servers/"] = httpGetTradeServerHandler(rs.apiSevers)
	m["/pservers/"] = httpGetTradeServerTcpHandler(rs.tradeServers)
	m["/login/"] = httpLoginHandler(rs.db, rs.sessionDuration)
	m["/validate/"] = httpAuthenticateHandler(rs.db, rs.checkExpires)
	m["/register/"] = httpUserCreationHandler(rs)
	m["/deleteaccount/"] = httpUserDeletionHandler(rs)
	m["/lookup/"] = httpGetUsernameHandler(rs)

	LOG.LogVerbose("Resolver HTTP server starting on %s:%d", CONN_HOST, rs.httpPort)
	go shared.NewHttpServer(rs.httpPort, m)

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, rs.tcpPort))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	LOG.LogVerbose("Resolver TCP server started on %s:%d", CONN_HOST, rs.tcpPort)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		LOG.CheckForError(err, false)
		// Handle connections in a new goroutine.
		go rs.onTradeServerConnection(conn)
	}

}

// Handles incoming requests.
func (rs *ResolverServer) onTradeServerConnection(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	var m shared.ResolverMessage
	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	LOG.CheckForError(err, false)
	_ = json.Unmarshal(buf[:n], &m)

	connectionHostPort := m.Payload
	h := GetHash(connectionHostPort)
	LOG.LogVerbose("ID for %s is %d", connectionHostPort, h)

	jId := shared.ResolverMessage{
		Type: shared.ID_ASSIGNMENT,
		Id:   h,
	}

	mId, _ := json.Marshal(jId)

	conn.Write(mId)
	exists := rs.servers.Get(h)

	rs.connectionMap.Put(h, conn)
	if !exists {
		LOG.LogVerbose("Adding %s to trade servers", connectionHostPort)
		rs.tradeServers.Put(h, connectionHostPort)
		rs.apiSevers.Append(fmt.Sprintf("%s:%d", strings.Split(connectionHostPort, ":")[0], m.Id))
		rs.servers.Add(h)
	} else {
		LOG.LogVerbose("%s has connected before", connectionHostPort)
	}
	conn.Close()
}

func GetHash(key string) uint32 {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

// ----------------------------------------------------
//                  HTTP handlers
// ----------------------------------------------------

func httpGetTradeServerHandler(apiSevers channelslice.ChannelSlice) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Print out all of the tradeservers
		for _, hostport := range apiSevers.GetStringList() {
			fmt.Fprintf(w, "%s,", hostport)
		}
	}
}

func httpGetTradeServerTcpHandler(tradeservers channelmap.ChannelMap) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Print out all of the tradeservers
		for id, hostport := range tradeservers.Raw() {
			fmt.Fprintf(w, "%s-%d,", hostport, id)
		}
	}
}

func httpGetUsernameHandler(rs *ResolverServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.FormValue("id")
		id, err := strconv.ParseUint(sid, 10, 32)
		userid := uint32(id)
		if err != nil {
			fmt.Fprintf(w, "Cannot parse item id")
			return
		}
		username, err := queryForUsername(userid, rs.db)
		LOG.CheckForError(err, false)
		if err != nil {
			fmt.Fprintf(w, "No such user")
			return
		}
		fmt.Fprintf(w, username)
	}
}

func httpLoginHandler(db *sql.DB, sessionDuration int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		pw := r.FormValue("password")
		LOG.LogVerbose("Login request for %s, %s", username, pw)
		u, err := queryForUser(username, db)
		if err != nil {
			LOG.CheckForError(err, false)
			fmt.Fprintf(w, "No such user")
		} else if u.password == pw {
			token := u.newToken(sessionDuration)
			_, err = db.Exec("UPDATE credentials SET token=? WHERE id=?", token, u.id)
			fmt.Fprintf(w, token)
		} else {
			fmt.Fprintf(w, "Incorrect Password")
		}
	}
}

func httpAuthenticateHandler(db *sql.DB, checkExpires bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		id := tokenToUserId(token, db, checkExpires)
		if id != -1 {
			fmt.Fprintf(w, fmt.Sprintf("%d", id))
		} else {
			fmt.Fprintf(w, "Invalid token")
		}
	}
}

func httpUserCreationHandler(rs *ResolverServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		pw := r.FormValue("password")
		_, err := queryForUser(username, rs.db)
		if err != nil {
			_, err := rs.db.Exec("INSERT INTO credentials (`username`, `password`, `token`) VALUES (?, ?, ?)", username, pw, "")
			LOG.CheckForError(err, false)
			fmt.Fprintf(w, "User %s Created!", username)
		} else {
			fmt.Fprintf(w, "User %s Exists", username)
		}
	}
}

func httpUserDeletionHandler(rs *ResolverServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		LOG.LogVerbose("Attempting to delete user")
		token := r.FormValue("token")
		id := tokenToUserId(token, rs.db, rs.checkExpires)
		if id != -1 {
			_, err := rs.db.Exec("DELETE FROM credentials WHERE id=?", id)
			LOG.CheckForError(err, false)
			fmt.Fprintf(w, "User Deleted")
		} else {
			fmt.Fprintf(w, "Invalid token")
		}
	}
}

// ----------------------------------------------------
//                HTTP helper functions
// ----------------------------------------------------

func queryForUser(name string, db *sql.DB) (*user, error) {
	row := db.QueryRow("select * from credentials where username=?", name)
	u := new(user)
	err := row.Scan(&u.id, &u.username, &u.password, &u.token)
	return u, err
}

func queryForUsername(id uint32, db *sql.DB) (string, error) {
	var name string
	row := db.QueryRow("select username from credentials where id=?", id)
	err := row.Scan(&name)
	return name, err
}

func (u *user) newToken(sessionDuration int) string {
	t := Token{
		u.username,
		time.Now().Add(time.Minute * time.Duration(sessionDuration)),
	}
	bt, err := json.Marshal(t)
	LOG.CheckForError(err, false)
	return base64.StdEncoding.EncodeToString(bt)
}

func tokenToUserId(encodedToken string, db *sql.DB, checkExpires bool) int {
	var token Token
	decodedToken, err := base64.StdEncoding.DecodeString(encodedToken)
	LOG.CheckForError(err, false)
	_ = json.Unmarshal([]byte(decodedToken), &token)

	type parsedData struct {
		id    int
		token string
	}

	row := db.QueryRow("select id,token from credentials where username=?", token.Username)
	st := new(parsedData)
	err = row.Scan(&st.id, &st.token)
	if err != nil {
		// User does not exist
		return -1
	}

	if st.token == encodedToken && (!checkExpires || time.Now().Before(token.Experation)) {
		return st.id
	}

	return -1
}
