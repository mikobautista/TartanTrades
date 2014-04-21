package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"net"
	"os"

	"github.com/mikobautista/tartantrades/channelmap"
	"github.com/mikobautista/tartantrades/channelvar"
	"github.com/mikobautista/tartantrades/logging"
)

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
)

var (
	LOG          = logging.NewLogger(true)
	tradeServers = channelmap.NewChannelMap()
	index        = channelvar.NewChannelVar()
)

func main() {
	var port = flag.Int("port", 1234, "Port to start resolver server on")
	flag.Parse()
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, *port))
	LOG.CheckForError(err, true)
	// Close the listener when the application closes.
	defer l.Close()
	LOG.LogVerbose("Resolver server started on %s:%d", CONN_HOST, *port)
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
	h := GetHash(conn.RemoteAddr().String())
	LOG.LogVerbose("Generating ID for %s", conn.RemoteAddr().String())
	sH := fmt.Sprintf("%d", h)
	tradeServers.Put(h, conn)
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
