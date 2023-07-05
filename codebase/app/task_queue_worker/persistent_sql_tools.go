package taskqueueworker

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
)

func generateAddColumnQuery(driverName, tableName, newColumnName, columnType string) (q []string) {
	switch driverName {
	case "postgres":
		q = []string{
			`SELECT column_name FROM information_schema.columns WHERE table_name='` + tableName + `' AND column_name='` + newColumnName + `'`,
			`ALTER TABLE ` + tableName + ` ADD COLUMN IF NOT EXISTS "` + newColumnName + `" ` + columnType,
		}

	case "sqlite3":
		q = []string{
			`SELECT name FROM pragma_table_info('` + tableName + `') WHERE name='` + newColumnName + `'`,
			`ALTER TABLE ` + tableName + ` ADD COLUMN "` + newColumnName + `" ` + columnType,
		}

	case "mysql":
		q = []string{
			"SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_NAME` = '" + tableName + "' AND `COLUMN_NAME` = '" + newColumnName + "'",
			`ALTER TABLE ` + tableName + " ADD COLUMN `" + newColumnName + "` " + columnType,
		}
	}

	return q
}

func (s *SQLPersistent) initTable(db *sql.DB) {
	var initTableQueries map[string]string

	switch s.driverName {
	case "postgres", "sqlite3":
		initTableQueries = map[string]string{
			jobModelName: `CREATE TABLE IF NOT EXISTS ` + jobModelName + ` (
				id VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
				task_name VARCHAR(255) NOT NULL DEFAULT '',
				arguments TEXT NOT NULL DEFAULT '',
				retries INTEGER NOT NULL DEFAULT 0,
				max_retry INTEGER NOT NULL DEFAULT 0,
				interval VARCHAR(255) NOT NULL DEFAULT '',
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				finished_at TIMESTAMPTZ NULL,
				status VARCHAR(255) NOT NULL DEFAULT '',
				error TEXT NOT NULL DEFAULT '',
				trace_id VARCHAR(255) NOT NULL DEFAULT '',
				current_progress INTEGER NOT NULL DEFAULT 0,
				max_progress INTEGER NOT NULL DEFAULT 0
			);
			CREATE INDEX IF NOT EXISTS idx_created_at ON ` + jobModelName + ` (created_at);
			CREATE INDEX IF NOT EXISTS idx_args_err ON ` + jobModelName + ` (arguments, error);
			CREATE INDEX IF NOT EXISTS idx_task_name_status_created_at ON ` + jobModelName + ` (task_name, status, created_at);
			CREATE INDEX IF NOT EXISTS idx_task_name ON ` + jobModelName + ` (task_name);
			CREATE INDEX IF NOT EXISTS idx_status ON ` + jobModelName + ` (status);
			CREATE INDEX IF NOT EXISTS idx_task_name_status ON ` + jobModelName + ` (task_name, status);`,

			jobSummaryModelName: `CREATE TABLE IF NOT EXISTS ` + jobSummaryModelName + ` (
				id VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
				success INTEGER NOT NULL DEFAULT 0,
				queueing INTEGER NOT NULL DEFAULT 0,
				retrying INTEGER NOT NULL DEFAULT 0,
				failure INTEGER NOT NULL DEFAULT 0,
				stopped INTEGER NOT NULL DEFAULT 0,
				is_loading BOOLEAN DEFAULT false,
				loading_message VARCHAR(255) NOT NULL DEFAULT ''
			);
			CREATE INDEX IF NOT EXISTS idx_task_name_summary ON ` + jobSummaryModelName + ` (id);`,

			jobHistoryModel: `CREATE TABLE IF NOT EXISTS ` + jobHistoryModel + ` (
				job_id VARCHAR(255) NOT NULL DEFAULT '',
				error_stack VARCHAR(255) NOT NULL DEFAULT '',
				status VARCHAR(255) NOT NULL DEFAULT '',
				error TEXT NOT NULL DEFAULT '',
				trace_id VARCHAR(255) NOT NULL DEFAULT '',
				start_at TIMESTAMP,
				end_at TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_job_id_history ON ` + jobHistoryModel + ` (job_id);
			CREATE INDEX IF NOT EXISTS idx_start_at_history ON ` + jobHistoryModel + ` (start_at);`,

			configurationModelName: `CREATE TABLE IF NOT EXISTS ` + configurationModelName + ` (
				key VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
				name VARCHAR(255) NOT NULL DEFAULT '',
				value VARCHAR(255) NOT NULL DEFAULT '',
				is_active BOOLEAN DEFAULT false
			);`,
		}

	case "mysql":
		initTableQueries = map[string]string{
			jobModelName: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n%s", jobModelName,
				"(`id` VARCHAR(255) PRIMARY KEY NOT NULL,",
				"`task_name` VARCHAR(255) NOT NULL,",
				"`arguments` TEXT NOT NULL,",
				"`retries` INTEGER NOT NULL,",
				"`max_retry` INTEGER NOT NULL,",
				"`interval` VARCHAR(255) NOT NULL,",
				"`created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,",
				"`updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,",
				"`finished_at` TIMESTAMP NULL,",
				"`status` VARCHAR(255) NOT NULL,",
				"`error` TEXT NOT NULL,",
				"`trace_id` VARCHAR(255) NOT NULL,",
				"`current_progress` INTEGER NOT NULL,",
				"`max_progress` INTEGER NOT NULL,",
				`INDEX (created_at),
				INDEX (arguments(255), error(255)),
				INDEX (task_name, status, created_at),
				INDEX (task_name),
				INDEX (status),
				INDEX (task_name, status)) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci;`,
			),
			jobSummaryModelName: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s %s %s %s\n%s", jobSummaryModelName+
				"(`id` VARCHAR(255) PRIMARY KEY NOT NULL,",
				"`success` INTEGER NOT NULL,",
				"`queueing` INTEGER NOT NULL,",
				"`retrying` INTEGER NOT NULL,",
				"`failure` INTEGER NOT NULL,",
				"`stopped` INTEGER NOT NULL,",
				"`is_loading` BOOLEAN DEFAULT false,",
				"`loading_message` VARCHAR(255) NOT NULL DEFAULT '',",
				`INDEX (id)) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci;`,
			),
			jobHistoryModel: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s %s %s %s\n%s", jobHistoryModel,
				"(`job_id` VARCHAR(255) NOT NULL,",
				"`error_stack` VARCHAR(255) NOT NULL,",
				"`status` VARCHAR(255) NOT NULL,",
				"`error` TEXT NOT NULL,",
				"`trace_id` VARCHAR(255) NOT NULL,",
				"`start_at` TIMESTAMP NULL,",
				"`end_at` TIMESTAMP NULL,",
				`INDEX (job_id),
				INDEX (start_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci;`,
			),
			configurationModelName: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s", configurationModelName,
				"(`key` VARCHAR(255) PRIMARY KEY NOT NULL,",
				"`name` VARCHAR(255) NOT NULL,",
				"`value` VARCHAR(255) NOT NULL,",
				"`is_active` BOOLEAN DEFAULT false) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci;",
			),
		}
	}

	for tableName, query := range initTableQueries {
		if err := s.checkExistingTable(db, tableName); err == nil {
			continue
		}
		if _, err := db.Exec(query); err != nil {
			panic(err)
		}
	}

	extraQueries := [][]string{
		generateAddColumnQuery(s.driverName, jobModelName, "result", "TEXT"),
		generateAddColumnQuery(s.driverName, jobHistoryModel, "result", "TEXT"),
	}
	for _, queries := range extraQueries {
		if queries[0] != "" {
			var columnName string
			if err := db.QueryRow(queries[0]).Scan(&columnName); err == nil {
				continue
			}
		}
		if _, err := db.Exec(queries[1]); err != nil {
			logger.LogE(err.Error())
		}
	}
}

func (s *SQLPersistent) checkExistingTable(db *sql.DB, tableName string) error {
	checkTableQuery := `SELECT EXISTS (SELECT * FROM ` + tableName + `);`
	if _, err := db.Exec(checkTableQuery); err != nil {
		return err
	}
	return nil
}

func (s *SQLPersistent) formatColumnName(columns ...string) string {
	switch s.driverName {
	case "mysql":
		for i, name := range columns {
			columns[i] = "`" + strings.TrimSpace(name) + "`"
		}
	}
	return strings.Join(columns, ",")
}

func (s *SQLPersistent) toMultiParamQuery(params []string) string {
	var in []string
	for _, param := range params {
		in = append(in, "'"+param+"'")
	}
	return " (" + strings.Join(in, ",") + ") "
}

func (s *SQLPersistent) parseNullTime(date time.Time) (t sql.NullTime) {
	t.Time = date
	t.Valid = !date.IsZero()
	return t
}

func (s *SQLPersistent) parseDateString(date string) (t sql.NullTime) {
	var err error
	for _, format := range []string{time.RFC3339, time.RFC3339Nano, candihelper.DateFormatYYYYMMDDHHmmss} {
		t.Time, err = time.Parse(format, date)
		if err == nil {
			break
		}
	}
	t.Time = candihelper.ToAsiaJakartaTime(t.Time)
	t.Valid = err == nil
	return t
}

func (s *SQLPersistent) parseDate(t time.Time) (res *string) {
	var date string
	switch s.driverName {
	case "postgres", "sqlite3":
		date = candihelper.ParseTimeToString(t, time.RFC3339Nano)
	case "mysql":
		date = candihelper.ParseTimeToString(t, candihelper.DateFormatYYYYMMDDHHmmss)
	}
	if date == "" {
		return nil
	}
	return &date
}

func (s *SQLPersistent) parameterize(lenCols int) (param string) {
	switch s.driverName {
	case "postgres":
		for i := 1; i <= lenCols; i++ {
			param += fmt.Sprintf("$%d", i)
			if i < lenCols {
				param += ","
			}
		}
		return param

	case "mysql", "sqlite3":
		for i := 0; i < lenCols; i++ {
			param += "?"
			if i < lenCols-1 {
				param += ","
			}
		}
		return param
	}
	return
}

func (s *SQLPersistent) parameterizeByColumnAndNumber(column string, number int) (param string) {
	switch s.driverName {
	case "postgres":
		return fmt.Sprintf(`"%s"=$%d`, column, number)

	case "mysql", "sqlite3":
		return fmt.Sprintf("%s=?", s.formatColumnName(column))
	}
	return
}

func (s *SQLPersistent) parameterizeForUpdate(columns ...string) (param string) {
	lenCol := len(columns)

	switch s.driverName {
	case "postgres":
		for i, column := range columns {
			param += fmt.Sprintf(`"%s"=$%d`, column, i+1)
			if i < lenCol-1 {
				param += ","
			}
		}
		return param

	case "mysql", "sqlite3":
		for i, column := range columns {
			param += fmt.Sprintf("%s=?", s.formatColumnName(column))
			if i < lenCol-1 {
				param += ","
			}
		}
		return param
	}

	return
}
