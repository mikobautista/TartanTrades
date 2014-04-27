package main

import (
	"flag"
	"github.com/mikobautista/tartantrades/server/resolver"
)

func main() {
	var tcpPort = flag.Int("tradeport", 1234, "Port to start resolver trade server on")
	var httpPort = flag.Int("httpport", 80, "Resolver http port")
	var tableName = flag.String("db", "accounts", "Database accounts table name")
	var dbUser = flag.String("db_user", "resolver", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")
	var sessionDuration = flag.Int("session_duration", 30, "Duration of a login session in minutes")
	var sessionExperation = flag.Bool("checkSessionExperation", true, "Sessions expire")
	flag.Parse()

	svr := resolver.NewResolverServer(
		*httpPort,
		*tcpPort,
		*tableName,
		*dbUser,
		*dbPw,
		*sessionDuration,
		*sessionExperation)

	svr.Start()

}
