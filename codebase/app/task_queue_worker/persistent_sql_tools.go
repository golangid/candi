package taskqueueworker

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
)

func (s *SQLPersistent) initTable(db *sql.DB) {
	var queries []string

	switch s.driverName {
	case "postgres", "sqlite3":
		queries = []string{
			`CREATE TABLE IF NOT EXISTS ` + jobModelName + ` (
				id VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
				task_name VARCHAR(255) NOT NULL DEFAULT '',
				arguments TEXT NOT NULL DEFAULT '',
				retries INTEGER NOT NULL DEFAULT 0,
				max_retry INTEGER NOT NULL DEFAULT 0,
				interval VARCHAR(255) NOT NULL DEFAULT '',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				finished_at TIMESTAMP NULL,
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
			`CREATE TABLE IF NOT EXISTS ` + jobSummaryModelName + ` (
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
			`CREATE TABLE IF NOT EXISTS task_queue_worker_job_histories (
				job_id VARCHAR(255) NOT NULL DEFAULT '',
				error_stack VARCHAR(255) NOT NULL DEFAULT '',
				status VARCHAR(255) NOT NULL DEFAULT '',
				error TEXT NOT NULL DEFAULT '',
				trace_id VARCHAR(255) NOT NULL DEFAULT '',
				start_at TIMESTAMP,
				end_at TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_job_id_history ON task_queue_worker_job_histories (job_id);
			CREATE INDEX IF NOT EXISTS idx_start_at_history ON task_queue_worker_job_histories (start_at);`,
			`CREATE TABLE IF NOT EXISTS ` + configurationModelName + ` (
				key VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
				name VARCHAR(255) NOT NULL DEFAULT '',
				value VARCHAR(255) NOT NULL DEFAULT '',
				is_active BOOLEAN DEFAULT false
			);`,
		}

	case "mysql":
		queries = []string{
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s\n%s", jobModelName,
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
				INDEX (task_name, status));`,
			),
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s %s %s %s\n%s", jobSummaryModelName+
				"(`id` VARCHAR(255) PRIMARY KEY NOT NULL,",
				"`success` INTEGER NOT NULL,",
				"`queueing` INTEGER NOT NULL,",
				"`retrying` INTEGER NOT NULL,",
				"`failure` INTEGER NOT NULL,",
				"`stopped` INTEGER NOT NULL,",
				"`is_loading` BOOLEAN DEFAULT false,",
				"`loading_message` VARCHAR(255) NOT NULL DEFAULT '',",
				`INDEX (id));`,
			),
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS task_queue_worker_job_histories %s %s %s %s %s %s %s\n%s",
				"(`job_id` VARCHAR(255) NOT NULL,",
				"`error_stack` VARCHAR(255) NOT NULL,",
				"`status` VARCHAR(255) NOT NULL,",
				"`error` TEXT NOT NULL,",
				"`trace_id` VARCHAR(255) NOT NULL,",
				"`start_at` TIMESTAMP NULL,",
				"`end_at` TIMESTAMP NULL,",
				`INDEX (job_id),
				INDEX (start_at));`,
			),
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s %s %s %s", configurationModelName,
				"(`key` VARCHAR(255) PRIMARY KEY NOT NULL,",
				"`name` VARCHAR(255) NOT NULL,",
				"`value` VARCHAR(255) NOT NULL,",
				"`is_active` BOOLEAN DEFAULT false);",
			),
		}
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			panic(err)
		}
	}
}

func (s *SQLPersistent) formatColumnName(c string) string {
	switch s.driverName {
	case "mysql":
		c = "`" + strings.TrimSpace(c) + "`"
	}
	return c
}

func (s *SQLPersistent) formatMultiColumnName(c string) string {
	switch s.driverName {
	case "mysql":
		splits := strings.Split(c, ",")
		for i, name := range splits {
			splits[i] = "`" + strings.TrimSpace(name) + "`"
		}
		return strings.Join(splits, ",")
	}
	return c
}

func (s *SQLPersistent) parseDateString(date string) (t time.Time) {
	switch s.driverName {
	case "postgres", "sqlite3":
		t, _ = time.Parse(time.RFC3339Nano, date)
	case "mysql":
		t, _ = time.Parse(candihelper.DateFormatYYYYMMDDHHmmss, date)
	}
	return candihelper.ToAsiaJakartaTime(t)
}

func (s *SQLPersistent) parseDate(t time.Time) (date string) {
	switch s.driverName {
	case "postgres", "sqlite3":
		date = candihelper.ParseTimeToString(t, time.RFC3339Nano)
	case "mysql":
		date = candihelper.ParseTimeToString(t, candihelper.DateFormatYYYYMMDDHHmmss)
	}
	return date
}
