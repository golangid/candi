package postgresworker

import (
	"database/sql"
	"fmt"
)

const (
	eventsConst              = "events"
	notifyEventFunctionQuery = `CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$

    DECLARE 
        data json;
        notification json;
    
    BEGIN
    
        -- Convert the old or new row to JSON, based on the kind of action.
        IF (TG_OP = 'DELETE') THEN
			data = json_build_object(
				'old', row_to_json(OLD)
			);
        ELSE
			data = json_build_object(
				'old', row_to_json(OLD),
				'new', row_to_json(NEW)
			);
        END IF;
        
        -- Construct the notification as a JSON string.
        notification = json_build_object(
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

// EventPayload event model
type EventPayload struct {
	Table  string `json:"table"`
	Action string `json:"action"`
}

func execCreateFunctionEventQuery(db *sql.DB) {
	query := `select pg_get_functiondef('notify_event()'::regprocedure);`
	var tmp string
	err := db.QueryRow(query).Scan(&tmp)
	if err != nil {
		stmt, _ := db.Prepare(notifyEventFunctionQuery)
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
		stmt, _ := db.Prepare(fmt.Sprintf(`CREATE TRIGGER %s_notify_event
		AFTER INSERT OR UPDATE OR DELETE ON %s
		FOR EACH ROW EXECUTE PROCEDURE notify_event();`, tableName, tableName))
		defer stmt.Close()

		if _, err = stmt.Exec(); err != nil {
			panic(fmt.Errorf("failed when create trigger for table %s: %s", tableName, err))
		}
	}
}
