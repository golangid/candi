package postgresworker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"github.com/lib/pq"
)

func getListener(dsn string) (*sql.DB, *pq.Listener) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, connection: %s`, err, candihelper.MaskingPasswordURL(dsn)))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, ping: %s`, err, candihelper.MaskingPasswordURL(dsn)))
	}

	listener := pq.NewListener(dsn, 10*time.Second, time.Minute, eventCallback)
	return db, listener
}

func eventCallback(ev pq.ListenerEventType, err error) {
	if err != nil {
		logger.LogRed(err.Error())
	}
}
