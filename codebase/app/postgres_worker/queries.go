package postgresworker

import (
	"database/sql"
	"fmt"

	"github.com/golangid/candi/logger"
)

const (
	eventsConst              = "events"
	notifyEventFunctionQuery = `CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$

	DECLARE 
		data json;
		notification json;
		
	BEGIN
		
		-- Convert the old or new row to JSON, based on the kind of action.
		data = json_build_object(
			'old', row_to_json(OLD),
			'new', row_to_json(NEW)
		);

		-- Construct the notification as a JSON string.
		notification = json_build_object(
						'event_id', md5(''||now()::text||random()::text),
						'table', TG_TABLE_NAME,
						'action', TG_OP,
						'data', data);
		
		-- Execute pg_notify(channel, notification)
		PERFORM pg_notify('events', notification::text);
		
		-- Result is ignored since this is an AFTER trigger
		RETURN NULL; 
	END;

$$ LANGUAGE plpgsql;`
)

type (
	// EventPayload event model
	EventPayload struct {
		EventID string           `json:"event_id"`
		Table   string           `json:"table"`
		Action  string           `json:"action"`
		Data    EventPayloadData `json:"data"`
	}
	// EventPayloadData event data
	EventPayloadData struct {
		Old interface{} `json:"old"`
		New interface{} `json:"new"`
	}
)

func execCreateFunctionEventQuery(db *sql.DB) {
	query := `select pg_get_functiondef('notify_event()'::regprocedure);`
	var tmp string
	err := db.QueryRow(query).Scan(&tmp)
	if err != nil {
		stmt, err := db.Prepare(notifyEventFunctionQuery)
		if err != nil {
			logger.LogYellow("Postgres Listener: warning, cannot create notify_event function. Error: " + err.Error())
			return
		}
		defer stmt.Close()

		if _, err = stmt.Exec(); err != nil {
			panic(fmt.Errorf("failed when create event function: %s", err))
		}
	}
}

func execTriggerQuery(db *sql.DB, tableName string) {
	query := `select event_object_table as table_name
		from information_schema.triggers
		where event_object_table=$1
		group by table_name`

	var existingTable string
	err := db.QueryRow(query, tableName).Scan(&existingTable)
	if err != nil {
		stmt, err := db.Prepare(`CREATE TRIGGER ` + tableName + `_notify_event
		AFTER INSERT OR UPDATE OR DELETE ON ` + tableName + `
		FOR EACH ROW EXECUTE PROCEDURE notify_event();`)
		if err != nil {
			logger.LogYellow("Postgres Listener: warning, cannot exec trigger for table " + tableName + ". Error: " + err.Error())
			return
		}
		defer stmt.Close()

		if _, err = stmt.Exec(); err != nil {
			panic(fmt.Errorf("failed when create trigger for table %s: %s", tableName, err))
		}
	}
}
