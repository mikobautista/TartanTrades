package main

import (
	"flag"
	"github.com/mikobautista/tartantrades/server/tradeserver"
)

func main() {
	var resolverhost = flag.String("resolverHost", "127.0.0.1", "Resolver host")
	var resolverHttpPort = flag.Int("resolverHttpPort", 80, "Resolver http port")
	var resolverTcpPort = flag.Int("resolverTcpPort", 1234, "Resolver tcp port")
	var httpPort = flag.Int("httpport", 80, "Trading http port")
	var tcpPort = flag.Int("tradeport", 1235, "Trading http port")
	var tableName = flag.String("db", "items", "Database accounts table name")
	var dbUser = flag.String("db_user", "trader", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")
	var tableDrop = flag.Bool("dropTableOnStart", false, "Drop table on start if it exists")
	var tableCreate = flag.Bool("createTableOnStart", false, "Create table on start")
	flag.Parse()
	svr := tradeserver.NewTradeServer(
		*resolverhost,
		*resolverTcpPort,
		*resolverHttpPort,
		*tcpPort,
		*httpPort,
		*tableName,
		*dbUser,
		*dbPw,
		*tableDrop,
		*tableCreate)

	svr.Start()
}
