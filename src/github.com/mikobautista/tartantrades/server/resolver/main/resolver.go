package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"time"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelslice"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"

	_ "github.com/ziutek/mymysql/godrv"
)

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG           = logging.NewLogger(true)
	connectionMap = channelmap.NewChannelMap()
	index         = channelvar.NewChannelVar()
	tradeServers  = channelslice.NewChannelSlice()
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

func main() {
	var tcpPort = flag.Int("tradeport", 1234, "Port to start resolver trade server on")
	var httpPort = flag.Int("httpport", 80, "Resolver http port")
	var tableName = flag.String("db", "accounts", "Database accounts table name")
	var dbUser = flag.String("db_user", "resolver", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")
	var sessionDuration = flag.Int("session_duration", 30, "Duration of a login session in minutes")
	var sessionExperation = flag.Bool("sessionExperation", true, "Sessions expire")
	flag.Parse()

	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", *tableName, *dbUser, *dbPw))
	LOG.CheckForError(err, true)
	defer db.Close()

	m := make(map[string]shared.HandlerType)
	m["/servers/"] = httpGetTradeServerHandler
	m["/login/"] = httpLoginHandler(db, *sessionDuration)
	m["/validate/"] = httpAuthenticateHandler(db, *sessionExperation)

	LOG.LogVerbose("Resolver HTTP server starting on %s:%d", CONN_HOST, *httpPort)
	go shared.NewHttpServer(*httpPort, m)

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, *tcpPort))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	LOG.LogVerbose("Resolver TCP server started on %s:%d", CONN_HOST, *tcpPort)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		LOG.CheckForError(err, false)
		// Handle connections in a new goroutine.
		go onTradeServerConnection(conn, db, *sessionExperation)
	}
}

// Handles incoming requests.
func onTradeServerConnection(conn net.Conn, db *sql.DB, sessionExperation bool) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	var m shared.ResolverMessage
	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	LOG.CheckForError(err, false)
	_ = json.Unmarshal(buf[:n], &m)
	LOG.LogVerbose("%s", string(buf))
	LOG.LogVerbose("%d - %s", int(m.Type), m.Payload)

	connectionHostPort := m.Payload
	h := GetHash(connectionHostPort)
	LOG.LogVerbose("ID for %s is %d", connectionHostPort, h)

	jId := shared.ResolverMessage{
		Type: shared.ID_ASSIGNMENT,
		Id:   h,
	}

	mId, _ := json.Marshal(jId)

	conn.Write(mId)

	// Notify trade servers of new server joining
	for _, v := range connectionMap.Raw() {
		joinMessage := shared.ResolverMessage{
			Type:    shared.TRADE_SERVER_JOIN,
			Payload: connectionHostPort,
		}
		marshalledMessage, _ := json.Marshal(joinMessage)
		v.(net.Conn).Write(marshalledMessage)
	}

	connectionMap.Put(h, conn)
	tradeServers.Append(connectionHostPort)

	// Send a hash back to person contacting us.
	go listenToTradeServer(conn, h, connectionHostPort, db, sessionExperation)
}

func GetHash(key string) uint32 {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

func listenToTradeServer(conn net.Conn, id uint32, connectionHostPort string, db *sql.DB, checkExpires bool) {
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOG.LogVerbose("TCP error for trade server id %d (%s), dropping...", id, connectionHostPort)
			conn.Close()
			connectionMap.Rem(id)
			tradeServers.Rem(connectionHostPort)

			//Notify Clients of drop
			for _, v := range connectionMap.Raw() {
				joinMessage := shared.ResolverMessage{
					Type:    shared.TRADE_SERVER_DROP,
					Payload: connectionHostPort,
				}
				marshalledMessage, _ := json.Marshal(joinMessage)
				v.(net.Conn).Write(marshalledMessage)
			}

			return
		}
		handleMessageFromTradeServer(buf[:n], id, conn, db, checkExpires)
	}
}

func handleMessageFromTradeServer(message []byte, id uint32, conn net.Conn, db *sql.DB, checkExpires bool) {
	var m shared.ResolverMessage
	_ = json.Unmarshal(message, &m)

	switch m.Type {
	case shared.AUTHENTICATION:
		conn.Write([]byte(fmt.Sprintf("%d", tokenToUserId(m.Payload, db, checkExpires))))
	}
}

// ----------------------------------------------------
//                  HTTP handlers
// ----------------------------------------------------

func httpGetTradeServerHandler(w http.ResponseWriter, r *http.Request) {
	// Print out all of the tradeservers
	for _, hostport := range tradeServers.GetStringList() {
		fmt.Fprintf(w, "%s\n", hostport)
	}
}

func httpLoginHandler(db *sql.DB, sessionDuration int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		pw := r.FormValue("password")
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

func queryForUser(name string, db *sql.DB) (*user, error) {
	row := db.QueryRow("select * from credentials where username=?", name)
	u := new(user)
	err := row.Scan(&u.id, &u.username, &u.password, &u.token)
	return u, err
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

	if st.token == encodedToken && (checkExpires && time.Now().Before(token.Experation)) {
		return st.id
	}

	return -1
}
