package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/mikobautista/tartantrades/logging"

	_ "github.com/ziutek/mymysql/godrv"
)

const (
	TABLE_NAME           = "`credentials`"
	CREATE_TABLE_COMMAND = "CREATE TABLE %s ( `id` int(11) NOT NULL AUTO_INCREMENT, `username` varchar(45) NOT NULL, `password` varchar(45) NOT NULL, `token` varchar(200) NOT NULL, PRIMARY KEY (`id`))"
)

var (
	LOG = logging.NewLogger(true)
)

func main() {
	var dbName = flag.String("db", "accounts", "Database accounts table name")
	var dbUser = flag.String("db_user", "root", "Database username")
	var dbPw = flag.String("db_pw", "password", "Database password")

	db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", *dbName, *dbUser, *dbPw))
	LOG.CheckForError(err, true)
	defer db.Close()

	LOG.LogVerbose("Dropping TABLE %s", TABLE_NAME)
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", TABLE_NAME))
	LOG.CheckForError(err, true)
	LOG.LogVerbose(CREATE_TABLE_COMMAND, TABLE_NAME)
	_, err = db.Exec(fmt.Sprintf(CREATE_TABLE_COMMAND, TABLE_NAME))
	LOG.CheckForError(err, true)
}
