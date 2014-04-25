package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/mikobautista/tartantrades/channelslice"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"

	_ "github.com/ziutek/mymysql/godrv"
)

type TradeServer struct {
	db               *sql.DB
	resolverhostport string
	tcpPort          int
	httpPort         int
	tableName        string
	dbUser           string
	dbPw             string
	thisHostPort     string
	index            channelvar.ChannelVar
	lowerBound       channelvar.ChannelVar
	tradeServers     channelslice.ChannelSlice
}

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG          = logging.NewLogger(true)
	index        = channelvar.NewChannelVar()
	lowerBound   = channelvar.NewChannelVar()
	tradeServers = channelslice.NewChannelSlice()
)

func NewTradeSever(
	resolverhostport string,
	tcpPort int,
	httpPort int,
	tableName string,
	dbUser string,
	dbPw string) *TradeServer {

	return &TradeServer{
		resolverhostport: resolverhostport,
		tcpPort:          tcpPort,
		httpPort:         httpPort,
		tableName:        tableName,
		dbUser:           dbUser,
		dbPw:             dbPw,
		thisHostPort:     fmt.Sprintf("%s:%d", CONN_HOST, tcpPort),
		index:            channelvar.NewChannelVar(),
		lowerBound:       channelvar.NewChannelVar(),
		tradeServers:     channelslice.NewChannelSlice(),
	}
}

func (ts *TradeServer) Start() {
	lowerBound.Set(0)

	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", ts.tableName, ts.dbUser, ts.dbPw))
	ts.db = db
	LOG.CheckForError(err, true)
	defer db.Close()

	// Launch HTTP server listening for clients
	m := make(map[string]shared.HandlerType)
	m["/availableblocks/"] = httpGetAvailableBlocks
	m["/sell/"] = httpGetAvailableBlocks
	LOG.LogVerbose("Starting HTTP server on port %d", ts.httpPort)
	go shared.NewHttpServer(ts.httpPort, m)

	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, ts.tcpPort))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	LOG.LogVerbose("Trade server TCP server started on %s", ts.thisHostPort)

	// (miko) Connect to the resolver via TCP
	LOG.LogVerbose("Connecting to resolver (%s)...", ts.resolverhostport)
	conn, err := net.Dial("tcp", ts.resolverhostport)
	LOG.CheckForError(err, true)

	connectMessage := shared.ResolverMessage{
		Type:    shared.CONNECT,
		Payload: ts.thisHostPort,
		Id:      uint32(ts.httpPort),
	}
	marshalledMessage, _ := json.Marshal(connectMessage)
	conn.Write(marshalledMessage)

	ts.listenToResolver(conn, func(id uint32) {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			LOG.CheckForError(err, false)

			// Handle connections in a new goroutine.
			go ts.onTradeServerConnection(conn)
		}
	})

}

func main() {
	var resolverhostport = flag.String("resolverhostport", "127.0.0.1:1234", "Resolver hostport")
	var httpPort = flag.Int("httpport", 80, "Trading http port")
	var tcpPort = flag.Int("tradeport", 1235, "Trading http port")
	var tableName = flag.String("db", "commits", "Database accounts table name")
	var dbUser = flag.String("db_user", "trader", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")
	flag.Parse()
	svr := NewTradeSever(
		*resolverhostport,
		*tcpPort,
		*httpPort,
		*tableName,
		*dbUser,
		*dbPw)

	svr.Start()
}

func (ts *TradeServer) listenToResolver(conn net.Conn, callback func(uint32)) {
	buf := make([]byte, 1024)
	hasIdAssigned := false
	var m shared.ResolverMessage
	for {
		n, err := conn.Read(buf)
		LOG.CheckForError(err, true)
		_ = json.Unmarshal(buf[:n], &m)

		switch m.Type {
		case shared.ID_ASSIGNMENT:
			if !hasIdAssigned {
				LOG.LogVerbose("Id of %d has been assigned, listening for trade server connections...", m.Id)
				go callback(m.Id)
			}
		case shared.TRADE_SERVER_JOIN:
			LOG.LogVerbose("Connecting to new trade server: %s", m.Payload)
			tradeServers.Append(m.Payload)
			conn, err := net.Dial("tcp", m.Payload)
			LOG.CheckForError(err, false)
			connectMessage := shared.TradeMessage{
				Type:    shared.WELCOME,
				Payload: ts.thisHostPort,
			}
			marshalledMessage, _ := json.Marshal(connectMessage)
			conn.Write(marshalledMessage)
			go ts.listenToTradeServer(conn, m.Payload)

		case shared.TRADE_SERVER_DROP:
			LOG.LogVerbose("Dropping trade server: %s", m.Payload)
			tradeServers.Rem(m.Payload)
		}
	}
}

func (ts *TradeServer) onTradeServerConnection(conn net.Conn) {
	buf := make([]byte, 1024)
	var m shared.TradeMessage
	n, err := conn.Read(buf)
	LOG.CheckForError(err, false)
	_ = json.Unmarshal(buf[:n], &m)
	switch m.Type {
	case shared.WELCOME:
		LOG.LogVerbose("Recieved Welcome from %s", m.Payload)
		go ts.listenToTradeServer(conn, m.Payload)
	}

}

func (ts *TradeServer) listenToTradeServer(conn net.Conn, otherHostPort string) {
	for {
		buf := make([]byte, 1024)
		var m shared.TradeMessage
		_, err := conn.Read(buf)
		if err != nil {
			LOG.LogError("Lost connection to %s", otherHostPort)
			return
		}
		switch m.Type {
		case shared.PREPARE:
		}
	}
}

// ----------------------------------------------------
//                  HTTP handlers
// ----------------------------------------------------

func httpGetAvailableBlocks(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

func httpMarkBlockForSale(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}
