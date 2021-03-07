package pinterest

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type UserConnect struct {
	host     string
	port     int
	user     string
	password string
	dbname   string
}

func (userInfo *UserConnect) UserInit(host, user, password, dbname string, port int) {
	userInfo.host = host
	userInfo.user = user
	userInfo.password = password
	userInfo.dbname = dbname
	userInfo.port = port
}

func (userInfo *UserConnect) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		userInfo.host, userInfo.port, userInfo.user, userInfo.password, userInfo.dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}
