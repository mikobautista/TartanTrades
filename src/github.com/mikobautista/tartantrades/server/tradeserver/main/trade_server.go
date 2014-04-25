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

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG          = logging.NewLogger(true)
	index        = channelvar.NewChannelVar()
	tradeServers = channelslice.NewChannelSlice()
)

func main() {
	var resolverhostport = flag.String("resolverhostport", "127.0.0.1:1234", "Resolver hostport")
	var httpPort = flag.Int("httpport", 80, "Trading http port")
	var tcpPort = flag.Int("tradeport", 1235, "Trading http port")
	var tableName = flag.String("db", "commits", "Database accounts table name")
	var dbUser = flag.String("db_user", "trader", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")
	flag.Parse()

	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", *tableName, *dbUser, *dbPw))
	LOG.CheckForError(err, true)
	defer db.Close()

	// Launch HTTP server listening for clients
	m := make(map[string]shared.HandlerType)
	m["/availableblocks/"] = httpGetAvailableBlocks
	m["/sell/"] = httpGetAvailableBlocks
	LOG.LogVerbose("Starting HTTP server on port %d", *httpPort)
	go shared.NewHttpServer(*httpPort, m)

	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, *tcpPort))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	thisHostPort := fmt.Sprintf("%s:%d", CONN_HOST, *tcpPort)
	LOG.LogVerbose("Trade server TCP server started on %s", thisHostPort)

	// (miko) Connect to the resolver via TCP
	LOG.LogVerbose("Connecting to resolver (%s)...", *resolverhostport)
	conn, err := net.Dial("tcp", *resolverhostport)
	LOG.CheckForError(err, true)

	connectMessage := shared.ResolverMessage{
		Type:    shared.CONNECT,
		Payload: thisHostPort,
	}
	marshalledMessage, _ := json.Marshal(connectMessage)
	conn.Write(marshalledMessage)

	listenToResolver(conn, thisHostPort, db, func(id uint32) {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			LOG.CheckForError(err, false)

			// Handle connections in a new goroutine.
			go onTradeServerConnection(conn, db)
		}
	})
}

func listenToResolver(conn net.Conn, thisHostPort string, db *sql.DB, callback func(uint32)) {
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
				Payload: thisHostPort,
			}
			marshalledMessage, _ := json.Marshal(connectMessage)
			conn.Write(marshalledMessage)
			go listenToTradeServer(conn, m.Payload, db)

		case shared.TRADE_SERVER_DROP:
			LOG.LogVerbose("Dropping trade server: %s", m.Payload)
			tradeServers.Rem(m.Payload)
		}
	}
}

func onTradeServerConnection(conn net.Conn, db *sql.DB) {
	buf := make([]byte, 1024)
	var m shared.TradeMessage
	n, err := conn.Read(buf)
	LOG.CheckForError(err, false)
	_ = json.Unmarshal(buf[:n], &m)
	switch m.Type {
	case shared.WELCOME:
		LOG.LogVerbose("Recieved Welcome from %s", m.Payload)
		go listenToTradeServer(conn, m.Payload, db)
	}

}

func listenToTradeServer(conn net.Conn, otherHostPort string, db *sql.DB) {
	for {
		buf := make([]byte, 1024)
		var m shared.ResolverMessage
		_, err := conn.Read(buf)
		if err != nil {
			LOG.LogError("Lost connection to %s", m.Payload)
			return
		}
		switch m.Type {
		//TODO: Fill in
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
