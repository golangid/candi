package postgresworker

import (
	"database/sql"
	"fmt"

	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/lib/pq"
)

// PostgresSource model
type PostgresSource struct {
	dsn         string
	name        string
	db          *sql.DB
	listener    *pq.Listener
	handlers    map[string]types.WorkerHandler
	workerIndex int
}

func (p *PostgresSource) execCreateFunctionEventQuery() error {
	query := `SELECT pg_get_functiondef('notify_event()'::regprocedure);`
	var tmp string
	err := p.db.QueryRow(query).Scan(&tmp)
	if err != nil {
		stmt, err := p.db.Prepare(notifyEventFunctionQuery)
		if err != nil {
			logger.LogYellow("Postgres Listener: warning, cannot create notify_event function. Error: " + err.Error())
			return err
		}
		defer stmt.Close()

		if _, err = stmt.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresSource) execTriggerQuery(tableName string) error {
	query := `SELECT event_object_table AS table_name
		FROM information_schema.triggers
		WHERE event_object_table=$1
		GROUP BY table_name`

	var existingTable string
	err := p.db.QueryRow(query, tableName).Scan(&existingTable)
	if err != nil {
		stmt, err := p.db.Prepare(`CREATE TRIGGER ` + tableName + `_notify_event
		AFTER INSERT OR UPDATE OR DELETE ON ` + tableName + `
		FOR EACH ROW EXECUTE PROCEDURE notify_event();`)
		if err != nil {
			logger.LogYellow("Postgres Listener: warning, cannot exec trigger for table " + tableName + ". Error: " + err.Error())
			return err
		}
		defer stmt.Close()

		if _, err = stmt.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresSource) findDetailData(tableName, id string) interface{} {
	rows, err := p.db.Query(`SELECT * FROM `+tableName+` WHERE id=$1`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil
	}

	results := make(map[string]interface{}, len(columns))
	if rows.Next() {
		values := make([]interface{}, len(columns))
		columnVals := make([]interface{}, len(columns))
		for i := range values {
			columnVals[i] = &values[i]
		}

		rows.Scan(columnVals...)
		for i, colName := range columns {
			results[colName] = values[i]
		}
	}

	return results
}

func (p *PostgresSource) getLogForSourceName() (sourceNameLog string) {
	if p.name != "" {
		sourceNameLog = fmt.Sprintf(" (source name: '%s')", p.name)
	}
	return
}
