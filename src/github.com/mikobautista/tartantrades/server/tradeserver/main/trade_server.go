package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"

	_ "github.com/ziutek/mymysql/godrv"
)

type TradeServer struct {
	// ---------------------
	//    Instance Details
	// ---------------------
	db               *sql.DB
	resolverhostport string
	tcpPort          int
	httpPort         int
	tableName        string
	dbUser           string
	dbPw             string
	thisHostPort     string
	tradeServers     channelmap.ChannelMap
	serverId         uint32

	// ---------------------
	//    Acceptor variables
	// ---------------------

	acceptedId    channelvar.ChannelVar
	promisedId    channelvar.ChannelVar
	acceptedValue channelvar.ChannelVar
	pendingAccept channelvar.ChannelVar

	// ---------------------
	//    Proposer variables
	// ---------------------
	SellChannel            chan sellInfo
	promiseRecievedChannel chan promise
}

type sellInfo struct {
	token string
	item  shared.Transaction
}

type promise struct {
	id       uint32
	willHold bool
	from     uint32
}

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG = logging.NewLogger(true)
)

func NewTradeSever(
	resolverhostport string,
	tcpPort int,
	httpPort int,
	tableName string,
	dbUser string,
	dbPw string) *TradeServer {

	svr := &TradeServer{
		resolverhostport:       resolverhostport,
		tcpPort:                tcpPort,
		httpPort:               httpPort,
		tableName:              tableName,
		dbUser:                 dbUser,
		dbPw:                   dbPw,
		thisHostPort:           fmt.Sprintf("%s:%d", CONN_HOST, tcpPort),
		acceptedId:             channelvar.NewChannelVar(),
		promisedId:             channelvar.NewChannelVar(),
		acceptedValue:          channelvar.NewChannelVar(),
		pendingAccept:          channelvar.NewChannelVar(),
		tradeServers:           channelmap.NewChannelMap(),
		SellChannel:            make(chan sellInfo, 200),
		promiseRecievedChannel: make(chan promise, 50),
	}
	svr.acceptedId.Set(uint32(0))
	svr.acceptedValue.Set(shared.Transaction{})
	return svr
}

func (ts *TradeServer) Start() {
	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", ts.tableName, ts.dbUser, ts.dbPw))
	ts.db = db
	LOG.CheckForError(err, true)
	defer db.Close()

	// Launch HTTP server listening for clients
	m := make(map[string]shared.HandlerType)
	m["/availableblocks/"] = httpGetAvailableBlocks
	m["/sell/"] = httpMarkBlockForSale(ts)
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

	go ts.listenForNewItems()

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
				ts.serverId = m.Id
				LOG.LogVerbose("Id of %d has been assigned, listening for trade server connections...", m.Id)
				go callback(m.Id)
			}
		case shared.TRADE_SERVER_JOIN:
			LOG.LogVerbose("Connecting to new trade server: %s", m.Payload)
			newConn, err := net.Dial("tcp", m.Payload)
			ts.tradeServers.Put(m.Payload, &newConn)
			LOG.CheckForError(err, false)
			connectMessage := shared.TradeMessage{
				Type:    shared.WELCOME,
				Payload: ts.thisHostPort,
			}
			marshalledMessage, _ := json.Marshal(connectMessage)
			newConn.Write(marshalledMessage)
			go ts.listenToTradeServer(newConn, m.Payload)

		case shared.TRADE_SERVER_DROP:
			LOG.LogVerbose("Dropping trade server: %s", m.Payload)
			ts.tradeServers.Rem(m.Payload)
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
		ts.tradeServers.Put(m.Payload, &conn)
		go ts.listenToTradeServer(conn, m.Payload)
	}

}

func (ts *TradeServer) listenToTradeServer(conn net.Conn, otherHostPort string) {
	for {
		buf := make([]byte, 1024)
		var m shared.TradeMessage
		n, err := conn.Read(buf)
		if err != nil {
			LOG.LogError("Lost connection to %s", otherHostPort)
			return
		}

		LOG.LogVerbose("Read %s from %s", string(buf), otherHostPort)
		_ = json.Unmarshal(buf[:n], &m)
		switch m.Type {
		case shared.PREPARE:
			// Duplicate Message
			if !ts.promisedId.IsNil() && m.ProposedId == ts.promisedId.Get().(uint32) {
				LOG.LogVerbose("Recieved duplicate perpare message, resending promise")
				ts.sendPromise(conn, shared.PROMISE, m.ProposedId, ts.acceptedId.Get().(uint32), ts.acceptedValue.Get().(shared.Transaction))
			} else if ts.promisedId.IsNil() || m.ProposedId > ts.promisedId.Get().(uint32) {
				LOG.LogVerbose("Recieved new prepare message from %s for id %d, Sending Promise...", otherHostPort, m.ProposedId)
				ts.promisedId.Set(m.ProposedId)
				ts.sendPromise(conn, shared.PROMISE, m.ProposedId, ts.acceptedId.Get().(uint32), ts.acceptedValue.Get().(shared.Transaction))
			} else {
				LOG.LogVerbose("Cannot promise %s for id %d", otherHostPort, m.ProposedId)
				ts.sendPromise(conn, shared.NPROMISE, m.ProposedId, ts.acceptedId.Get().(uint32), ts.acceptedValue.Get().(shared.Transaction))
			}
		case shared.PROMISE:
			ts.promiseRecievedChannel <- promise{m.PromisedId, true, m.FromNodeId}
		case shared.NPROMISE:
			ts.promiseRecievedChannel <- promise{m.PromisedId, true, m.FromNodeId}
		case shared.ACCEPT:
			if ts.promisedId.IsNil() || m.ProposedId >= ts.promisedId.Get().(uint32) {
				LOG.LogVerbose("Accepting transaction")
				ts.promisedId.Set(m.ProposedId)
				ts.acceptedId.Set(m.ProposedId)
				ts.applyTransaction(m.AcceptedValue)
				ts.sendAccepted(conn, m.ProposedId, m.AcceptedValue)
			} else {
				LOG.LogVerbose("Failed to accept Transaction")
			}
		}
	}
}

func (ts *TradeServer) applyTransaction(t shared.Transaction) {

}

func (ts *TradeServer) listenForNewItems() {
	for {
		select {
		case item := <-ts.SellChannel:
			//TODO remove this later
			_ = item
			LOG.LogVerbose("Selling item %s", item.item)
			prepareRequestId := ts.acceptedId.Get().(uint32) + 1
			ts.sendPrepare(prepareRequestId)
			numberOfParticipants := len(ts.tradeServers.Raw())
			quorum := numberOfParticipants / 2
			acceptCount := 0
			rejectCount := 0
			var seen = make([]uint32, numberOfParticipants)
			isSeen := func(i uint32) bool {
				for _, v := range seen {
					if i == v {
						return true
					}
				}
				return false
			}

			for {
				select {
				case p := <-ts.promiseRecievedChannel:
					// Ignore old promises
					if p.id == prepareRequestId {
						if !isSeen(p.from) {
							seen = append(seen, p.from)
							if p.willHold {
								acceptCount++
							} else {
								rejectCount++
							}
							// Check to see if we're done
							if acceptCount+rejectCount >= numberOfParticipants {
								// If majority promised
								if acceptCount >= quorum {
									LOG.LogVerbose("Majority accepted proposal (%d to %d)", acceptCount, rejectCount)
									ts.sendAccept(p.id, item.item)
									break

								} else {
									LOG.LogVerbose("Majority rejected proposal for %s (%d to %d)", item.item, acceptCount, rejectCount)
									// If fail,try again later
									ts.SellChannel <- item
								}
							}
						}
					}
				}
			}
		}
	}
}

func (ts *TradeServer) sendPrepare(proposalId uint32) {
	prepareMessage := shared.TradeMessage{
		Type:       shared.PREPARE,
		ProposedId: proposalId,
	}
	ts.blast(prepareMessage)
}

func (ts *TradeServer) blast(message shared.TradeMessage) {
	for _, c := range ts.tradeServers.Raw() {
		marshalledMessage, _ := json.Marshal(message)
		_, err := (*(c.(*net.Conn))).Write(marshalledMessage)
		LOG.CheckForError(err, false)
	}
}

func (ts *TradeServer) sendPromise(proposerConn net.Conn, promiseType shared.TradeMessageType, promisedId uint32, acceptedId uint32, acceptedValue shared.Transaction) {
	promiseMessage := shared.TradeMessage{
		Type:          promiseType,
		FromNodeId:    ts.serverId,
		PromisedId:    promisedId,
		AcceptedId:    acceptedId,
		AcceptedValue: acceptedValue,
	}
	marshalledMessage, _ := json.Marshal(promiseMessage)
	proposerConn.Write(marshalledMessage)
}

func (ts *TradeServer) sendAccept(proposalId uint32, acceptedValue shared.Transaction) {
	acceptMessage := shared.TradeMessage{
		Type:          shared.ACCEPT,
		ProposedId:    proposalId,
		AcceptedValue: acceptedValue,
	}
	ts.blast(acceptMessage)
}

func (ts *TradeServer) sendAccepted(proposerConn net.Conn, proposalId uint32, acceptedValue shared.Transaction) {
	promiseMessage := shared.TradeMessage{
		Type:          shared.ACCEPTED,
		ProposedId:    proposalId,
		AcceptedValue: acceptedValue,
	}
	marshalledMessage, _ := json.Marshal(promiseMessage)
	proposerConn.Write(marshalledMessage)

}

// ----------------------------------------------------
//                  HTTP handlers
// ----------------------------------------------------

func httpGetAvailableBlocks(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

func httpMarkBlockForSale(ts *TradeServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		ts.SellChannel <- sellInfo{token, shared.Transaction{
			Type:  shared.SELL,
			Token: token,
		}}
	}
}
