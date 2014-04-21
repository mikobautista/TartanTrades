package main

import (
	"flag"
	"net"
	"os"
	"strconv"

	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"
)

var (
	LOG = logging.NewLogger(true)
)

type tradeServer struct {
	port                 int
	nodeID               int
	masterServerHostPort string
	clientListener       net.Listener
	resolverConn         net.Conn
	registerChan         chan net.Conn
}

func main() {
	var resolverhostport = flag.String("resolverhostport", "127.0.0.1:1234", "Port to start resolver trade server on")
	var httpPort = flag.Int("httpport", 80, "Port to start resolver http server on")
	flag.Parse()

	if len(os.Args) != 3 {
		LOG.LogVerbose("tradeserver usage: ./main -resolverhostport=<hostport> -httpport=<port>")
		return
	}

	// (miko) Connect to the resolver via TCP
	conn, err := net.Dial("tcp", *resolverhostport)
	LOG.CheckForError(err, true)
	conn.Write([]byte("connect"))

	// (miko) Wait for acknowledgement from resolver
	ipAddrHash, err := getIntFromConnection(conn)
	LOG.CheckForError(err, true)
	LOG.LogVerbose("Obtained hash from server %d", ipAddrHash)

	// Launch HTTP server listening for clients
	m := make(map[string]shared.HandlerType)
	// TODO: map paths to functions
	shared.NewHttpServer(*httpPort, m)
}

// Handles reading the messages from resolver
func (ts *tradeServer) readHandler() {
	for {
		paxosNum, err := getIntFromConnection(ts.resolverConn)
		if err != nil {
			// TODO: resolver disconnected
		} else {
			LOG.LogVerbose("New Paxos Num: %d", paxosNum)
			// TODO: resolver sent updated Paxos count
		}
	}
}

func (ts *tradeServer) serverLoop() {
	for {
		select {
		// TODO: do something
		}
	}
}

func getIntFromConnection(conn net.Conn) (int, error) {
	var buf [1024]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(string(buf[0:n]))
	LOG.CheckForError(err, true)
	return num, nil
}
