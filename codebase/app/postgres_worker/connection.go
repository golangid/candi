package postgresworker

import (
	"database/sql"
	"fmt"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/database"
	"github.com/golangid/candi/logger"
	"github.com/lib/pq"
)

func getListener(source string, opts *option) (*sql.DB, *pq.Listener) {
	driverName, dsn := database.ParseSQLDSN(source)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, connection: %s`, err, candihelper.MaskingPasswordURL(dsn)))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf(`[POSTGRES-LISTENER] ERROR: %v, ping: %s`, err, candihelper.MaskingPasswordURL(dsn)))
	}

	if opts.dbOption != nil {
		opts.dbOption(db)
	}

	ec := &eventCallback{onErrorFunc: opts.onErrorConnectionFunc}
	listener := pq.NewListener(dsn, opts.minReconnectInterval, opts.maxReconnectInterval, ec.onEvent)
	return db, listener
}

type eventCallback struct {
	onErrorFunc func(error)
}

func (e *eventCallback) onEvent(ev pq.ListenerEventType, err error) {
	switch ev {
	case pq.ListenerEventConnected, pq.ListenerEventReconnected:
		logger.LogYellow("[POSTGRES-LISTENER] Ready to receive event")
	}
	if err != nil {
		logger.LogRed("[POSTGRES-LISTENER] ERROR when listening: " + err.Error())
		if e.onErrorFunc != nil {
			e.onErrorFunc(err)
		}
	}
}
