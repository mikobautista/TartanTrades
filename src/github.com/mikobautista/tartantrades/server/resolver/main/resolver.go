package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"os"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelslice"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/server/shared"
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

func main() {
	var tcpPort = flag.Int("tradeport", 1234, "Port to start resolver trade server on")
	var httpPort = flag.Int("httpport", 80, "Port to start resolver http server on")
	flag.Parse()

	m := make(map[string]shared.HandlerType)
	m["/servers/"] = httpGetTradeServerHandler

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
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	LOG.CheckForError(err, false)
	connectionHostPort := conn.RemoteAddr().String()
	h := GetHash(connectionHostPort)
	LOG.LogVerbose("Generating ID for %s", connectionHostPort)
	sH := fmt.Sprintf("%d", h)
	connectionMap.Put(h, conn)
	tradeServers.Append(connectionHostPort)
	// Send a hash back to person contacting us.
	conn.Write([]byte(sH))
	// Close the connection when you're done with it.
	conn.Close()
}

func GetHash(key string) uint32 {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

func httpGetTradeServerHandler(w http.ResponseWriter, r *http.Request) {
	// Print out all of the tradeservers
	for _, hostport := range tradeServers.GetStringList() {
		fmt.Fprintf(w, "%s\n", hostport)
	}
}
