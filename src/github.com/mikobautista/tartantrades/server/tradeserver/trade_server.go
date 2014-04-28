package tradeserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"
	"github.com/mikobautista/tartantrades/set"

	_ "github.com/ziutek/mymysql/godrv"
)

type TradeServer struct {
	// ---------------------
	//    Instance Details
	// ---------------------
	db                   *sql.DB
	resolverTcpHostPort  string
	resolverHttpHostPort string
	tcpPort              int
	httpPort             int
	databaseName         string
	dbUser               string
	dbPw                 string
	commitTableName      string
	buyTableName         string
	dropTableOnStart     bool
	createTableOnStart   bool
	thisHostPort         string
	serverId             uint32
	// hostport:string -> conn net.conn
	tradeServers channelmap.ChannelMap

	// ---------------------
	//    Acceptor variables
	// ---------------------
	// uint32
	acceptedId channelvar.ChannelVar
	// uint32
	promisedId channelvar.ChannelVar
	// Transaction
	acceptedValue channelvar.ChannelVar

	// ---------------------
	//    Proposer variables
	// ---------------------
	TransactionChannel     chan transactionInfo
	promiseRecievedChannel chan promise

	// ---------------------
	//        Data
	// --------------------
	// coordinate -> seller id
}

type transactionInfo struct {
	token string
	item  shared.Transaction
}

type availableItem struct {
	Commit_id uint32
	X         string
	Y         string
	Seller_id uint32
}

type soldItem struct {
	Commit_id uint32
	BuyerId   uint32
}

type promise struct {
	id       uint32
	willHold bool
	from     uint32
}

const (
	CONN_HOST                     = "localhost"
	CONN_TYPE                     = "tcp"
	CREATE_COMMIT_TABLE_STATEMENT = "CREATE TABLE `?` ( `commit_id` int(11) NOT NULL AUTO_INCREMENT, `x` varchar(45) NOT NULL, `y` varchar(45) NOT NULL, `seller_id` INT NOT NULL, PRIMARY KEY (`commit_id`));"
	CREATE_BUY_TABLE_STATEMENT    = "CREATE TABLE `?` ( `commit_sold` INT NOT NULL, `buyer` INT NOT NULL, PRIMARY KEY (`commit_sold`)); "
)

var (
	LOG = logging.NewLogger(true)
)

func NewTradeServer(
	resolverhost string,
	resolverTcpPort int,
	resolverHttpPort int,
	tcpPort int,
	httpPort int,
	databaseName string,
	dbUser string,
	dbPw string,
	dropTableOnStart bool,
	createTableOnStart bool) *TradeServer {

	svr := &TradeServer{
		resolverTcpHostPort:    fmt.Sprintf("%s:%d", resolverhost, resolverTcpPort),
		resolverHttpHostPort:   fmt.Sprintf("%s:%d", resolverhost, resolverHttpPort),
		tcpPort:                tcpPort,
		httpPort:               httpPort,
		databaseName:           databaseName,
		dbUser:                 dbUser,
		dbPw:                   dbPw,
		thisHostPort:           fmt.Sprintf("%s:%d", CONN_HOST, tcpPort),
		acceptedId:             channelvar.NewChannelVar(),
		promisedId:             channelvar.NewChannelVar(),
		acceptedValue:          channelvar.NewChannelVar(),
		tradeServers:           channelmap.NewChannelMap(),
		TransactionChannel:     make(chan transactionInfo, 200),
		promiseRecievedChannel: make(chan promise, 50),
		dropTableOnStart:       dropTableOnStart,
		createTableOnStart:     createTableOnStart,
	}
	svr.acceptedId.Set(uint32(0))
	svr.acceptedValue.Set(shared.Transaction{})
	return svr
}

func (ts *TradeServer) Start() {
	LOG.LogVerbose("Establishing Connection to Database...")
	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", ts.databaseName, ts.dbUser, ts.dbPw))
	ts.db = db
	LOG.CheckForError(err, true)

	defer db.Close()

	// Launch HTTP server listening for clients
	m := make(map[string]shared.HandlerType)
	m["/availableblocks/"] = httpGetAvailableBlocks(ts)
	m["/sell/"] = httpMarkBlockForSale(ts)
	m["/buy/"] = httpPurchaseHandler(ts)
	m["/stop/"] = httpStopHandler
	LOG.LogVerbose("Starting HTTP server on port %d", ts.httpPort)
	go shared.NewHttpServer(ts.httpPort, m)

	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, ts.tcpPort))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	LOG.LogVerbose("Trade server TCP server started on %s", ts.thisHostPort)

	// (miko) Connect to the resolver via TCP
	LOG.LogVerbose("Connecting to resolver (%s)...", ts.resolverTcpHostPort)
	conn, err := net.Dial("tcp", ts.resolverTcpHostPort)
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
		ts.commitTableName = fmt.Sprintf("commits_%d", id)
		ts.buyTableName = fmt.Sprintf("purchases_%d", id)
		if ts.dropTableOnStart {
			LOG.LogVerbose("Dropping tables %s and %s", ts.commitTableName, ts.buyTableName)
			_, err = db.Exec("DROP TABLE IF EXISTS `?`", ts.commitTableName)
			_, err = db.Exec("DROP TABLE IF EXISTS `?`", ts.buyTableName)
			LOG.CheckForError(err, true)
		}
		if ts.createTableOnStart {
			LOG.LogVerbose("Creating tables %s and %s", ts.commitTableName, ts.buyTableName)
			_, err = db.Exec(CREATE_COMMIT_TABLE_STATEMENT, ts.commitTableName)
			_, err = db.Exec(CREATE_BUY_TABLE_STATEMENT, ts.buyTableName)
			LOG.CheckForError(err, true)
		}

		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			LOG.CheckForError(err, false)

			// Handle connections in a new goroutine.
			go ts.onTradeServerConnection(conn)
		}
	})
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
		if len(ts.tradeServers.Raw()) == 0 {
			ts.checkRecovery(conn)
		}
		ts.tradeServers.Put(m.Payload, &conn)
		go ts.listenToTradeServer(conn, m.Payload)
	}
}

func (ts *TradeServer) checkRecovery(otherTradeServer net.Conn) {
	LOG.LogVerbose("Checking to see if recovery is necessary...")
	recoveryCheckMessage := shared.TradeMessage{
		Type:       shared.RECOVER_CHECK,
		AcceptedId: ts.acceptedId.Get().(uint32),
	}
	marshalledMessage, _ := json.Marshal(recoveryCheckMessage)
	otherTradeServer.Write(marshalledMessage)
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
				ts.commit(m.ProposedId, m.AcceptedValue)
				ts.sendAccepted(conn, m.ProposedId, m.AcceptedValue)
			} else {
				LOG.LogVerbose("Failed to accept Transaction")
			}
		case shared.RECOVER_CHECK:
			if m.AcceptedId < ts.acceptedId.Get().(uint32) {
				marshalledData, _ := json.Marshal(ts.getTransactionsAfterCommit(m.AcceptedId))
				LOG.LogVerbose("%s Needs to recover, sending data (%s)...", otherHostPort, string(marshalledData))
				recoverMessage := shared.TradeMessage{
					Type:       shared.RECOVER_NECESSARY,
					Payload:    string(marshalledData),
					AcceptedId: ts.acceptedId.Get().(uint32),
				}
				marshalledMessage, _ := json.Marshal(recoverMessage)
				conn.Write(marshalledMessage)
			} else {
				LOG.LogVerbose("%s Does not need to recover...", otherHostPort)
			}
		case shared.RECOVER_NECESSARY:
			var missedCommits []*availableItem
			_ = json.Unmarshal([]byte(m.Payload), &missedCommits)
			ts.recoverTransactions(missedCommits)
			ts.acceptedId.Set(m.AcceptedId)
			LOG.LogVerbose("Recovery is necessary - recovering data from %s (now on accepted id %d)", otherHostPort, m.AcceptedId)
		}
	}
}

func (ts *TradeServer) getTransactionsAfterCommit(commitId uint32) []*availableItem {
	return ts.getAllItemsWithQuery(ts.db.Query("SELECT * FROM `?` WHERE `commit_id` > ? ORDER BY commit_id ASC", ts.commitTableName, commitId))
}

func (ts *TradeServer) recoverTransactions(items []*availableItem) {
	for _, item := range items {
		LOG.LogVerbose("Marking (%s,%s)>%d as available", item.X, item.Y, item.Seller_id)
		ts.markItemAsAvailable(item.X, item.Y, item.Seller_id)
	}
}

func (ts *TradeServer) markItemAsAvailable(x string, y string, sellerId uint32) {
	_, err := ts.db.Exec("INSERT INTO `?` (`commit_id`, `x`, `y`, `seller_id`) VALUES (NULL, ?, ?, ?)", ts.commitTableName, x, y, sellerId)
	LOG.CheckForError(err, true)
}

func (ts *TradeServer) markItemAsSold(commitId uint32, buyer uint32) {
	//INSERT INTO `items`.`'purchases_2698766122'` (`commit_sold`, `buyer`) VALUES ('3', '1');
	_, err := ts.db.Exec("INSERT INTO `?` (`commit_sold`, `buyer`) VALUES (?, ?)", ts.buyTableName, commitId, buyer)
	LOG.CheckForError(err, false)
}

func (ts *TradeServer) applyTransaction(t shared.Transaction, commitId uint32) {
	switch t.Type {
	case shared.SELL:
		ts.markItemAsAvailable(t.X, t.Y, t.Id)
	case shared.PURCHASE:
		ts.markItemAsSold(t.Commit, t.To)
	}
}

func (ts *TradeServer) listenForNewItems() {
	for {
		select {
		case item := <-ts.TransactionChannel:
			prepareRequestId := ts.acceptedId.Get().(uint32) + 1
			ts.sendPrepare(prepareRequestId)
			numberOfParticipants := len(ts.tradeServers.Raw())
			quorum := numberOfParticipants / 2
			acceptCount := 0
			rejectCount := 0
			waiting := true
			var seen = make([]uint32, numberOfParticipants)
			isSeen := func(i uint32) bool {
				for _, v := range seen {
					if i == v {
						return true
					}
				}
				return false
			}
			if numberOfParticipants == 0 {
				LOG.LogVerbose("Only Server, commiting transaction.")
				ts.commit(prepareRequestId, item.item)
			} else {
				for waiting {
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
										ts.commit(prepareRequestId, item.item)
										waiting = false

									} else {
										LOG.LogVerbose("Majority rejected proposal for %s (%d to %d)", item.item, acceptCount, rejectCount)
										// If fail,try again later
										ts.TransactionChannel <- item
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (ts *TradeServer) commit(id uint32, item shared.Transaction) {
	ts.promisedId.Set(id)
	ts.acceptedId.Set(id)
	ts.applyTransaction(item, id)
}

func (ts *TradeServer) sendPrepare(proposalId uint32) {
	prepareMessage := shared.TradeMessage{
		Type:       shared.PREPARE,
		ProposedId: proposalId,
	}
	ts.blast(prepareMessage)
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

func (ts *TradeServer) blast(message shared.TradeMessage) {
	for _, c := range ts.tradeServers.Raw() {
		marshalledMessage, _ := json.Marshal(message)
		_, err := (*(c.(*net.Conn))).Write(marshalledMessage)
		LOG.CheckForError(err, false)
	}
}

// ----------------------------------------------------
//                  HTTP handlers
// ----------------------------------------------------

func httpGetAvailableBlocks(ts *TradeServer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		items := ts.getAvailableItems()
		s := set.NewUint32Set()
		for _, i := range ts.getAllSoldItems() {
			s.Add(i.Commit_id)
		}
		for _, item := range items {
			if !s.Get(item.Commit_id) {
				fmt.Fprintf(w, "%s,%s>%d:%d;", item.X, item.Y, item.Seller_id, item.Commit_id)
			}
		}
	}
}

func (ts *TradeServer) getAllItemsWithQuery(rows *sql.Rows, err error) []*availableItem {
	LOG.CheckForError(err, false)
	returnMe := make([]*availableItem, 0)
	defer rows.Close()
	for rows.Next() {
		var item availableItem
		err := rows.Scan(&item.Commit_id, &item.X, &item.Y, &item.Seller_id)
		LOG.CheckForError(err, false)
		returnMe = append(returnMe, &item)
	}
	err = rows.Err()
	LOG.CheckForError(err, false)
	return returnMe
}

func (ts *TradeServer) getSoldItemsWithQuery(rows *sql.Rows, err error) []*soldItem {
	LOG.CheckForError(err, false)
	returnMe := make([]*soldItem, 0)
	defer rows.Close()
	for rows.Next() {
		var item soldItem
		err := rows.Scan(&item.Commit_id, &item.BuyerId)
		LOG.CheckForError(err, false)
		returnMe = append(returnMe, &item)
	}
	err = rows.Err()
	LOG.CheckForError(err, false)
	return returnMe
}

func (ts *TradeServer) getAvailableItems() []*availableItem {
	return ts.getAllItemsWithQuery(ts.db.Query("select * from `?`", ts.commitTableName))
}

func (ts *TradeServer) getAllSoldItems() []*soldItem {
	return ts.getSoldItemsWithQuery(ts.db.Query("select * from `?`", ts.buyTableName))
}

func httpMarkBlockForSale(ts *TradeServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		x := r.FormValue("x")
		y := r.FormValue("y")
		i, err := ts.tokenToUserId(token)
		if err != nil {
			fmt.Fprintf(w, "Invalid token")
			return
		}

		ts.TransactionChannel <- transactionInfo{token, shared.Transaction{
			Type: shared.SELL,
			Id:   uint32(i),
			X:    x,
			Y:    y,
		}}
	}
}

func (ts *TradeServer) tokenToUserId(token string) (uint64, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/validate/?token=%s", ts.resolverHttpHostPort, token))
	LOG.CheckForError(err, false)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return strconv.ParseUint(string(body), 10, 32)
}

func httpPurchaseHandler(ts *TradeServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		sItem := r.FormValue("item")
		userid, err := ts.tokenToUserId(token)
		if err != nil {
			fmt.Fprintf(w, "Invalid token")
			return
		}
		id, err := strconv.ParseUint(sItem, 10, 32)
		purchaseid := uint32(id)

		if err != nil {
			fmt.Fprintf(w, "Cannot parse item id")
			return
		}
		if purchaseid > ts.acceptedId.Get().(uint32) {
			fmt.Fprintf(w, "Invalid Item")
			return
		}

		s := set.NewUint32Set()
		for _, i := range ts.getAllSoldItems() {
			s.Add(i.Commit_id)
		}
		if s.Get(purchaseid) {
			fmt.Fprintf(w, "Item has already been purchased")
			return
		}

		ts.TransactionChannel <- transactionInfo{token, shared.Transaction{
			Type:   shared.PURCHASE,
			Commit: purchaseid,
			To:     uint32(userid),
		}}
	}
}

func httpStopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Stopping trade server in 5 seconds")
	go func() {
		time.Sleep(time.Second * 5)
		os.Exit(10)
	}()
}
