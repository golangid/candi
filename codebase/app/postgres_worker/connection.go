package postgresworker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
)

func getListener() (*sql.DB, *pq.Listener) {

	db, err := sql.Open("postgres", env.BaseEnv().DbSQLWriteDSN)
	if err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, connection: %s`, err, candihelper.MaskingPasswordURL(env.BaseEnv().DbSQLWriteDSN)))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, ping: %s`, err, candihelper.MaskingPasswordURL(env.BaseEnv().DbSQLWriteDSN)))
	}

	listener := pq.NewListener(env.BaseEnv().DbSQLWriteDSN, 10*time.Second, time.Minute, eventCallback)
	return db, listener
}

func eventCallback(ev pq.ListenerEventType, err error) {
	if err != nil {
		logger.LogRed(err.Error())
	}
}
