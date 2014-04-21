package main

import (
    "fmt"
    "github.com/mikobautista/tartantrades/logging"
    "net"
    "net/http"
    "os"
    "strconv"
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
    if len(os.Args) != 3 {
        LOG.LogVerbose("tradeserver usage: ./main <masterServerHostPort> <port>")
        return
    }
    masterServerHostPort := os.Args[1]
    port, err := strconv.Atoi(os.Args[2])
    LOG.CheckForError(err, true)

    // (miko) Connect to the resolver via TCP
    conn, err := net.Dial("tcp", masterServerHostPort)
    LOG.CheckForError(err, true)
    conn.Write([]byte("hi!"))

    // (miko) Wait for acknowledgement from resolver
    ipAddrHash, err := getIntFromConnection(conn)
    LOG.CheckForError(err, true)
    LOG.LogVerbose("Obtained hash from server %d", ipAddrHash)

    // (miko) Create the server socket that will listen for clients
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    LOG.CheckForError(err, true)
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

    // (miko) Keep track of args that were passed in
    ts := &tradeServer{
        port:                 port,
        clientListener:       ln,
        resolverConn:         conn,
        nodeID:               ipAddrHash,
        masterServerHostPort: masterServerHostPort,
        registerChan:         make(chan net.Conn),
    }

    // Launch the server
    LOG.LogVerbose("Listening on port %d", port)
    go ts.readHandler()
    go ts.serverLoop()
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
