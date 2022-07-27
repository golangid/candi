package taskqueueworker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"github.com/google/uuid"
)

type (
	SQLPersistent struct {
		db            *sql.DB
		summary       Summary
		queryReplacer *strings.Replacer
	}
)

// NewSQLPersistent init new persistent SQL
func NewSQLPersistent(db *sql.DB) *SQLPersistent {

	// init jobs table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ` + jobModelName + ` (
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
		error VARCHAR(255) NOT NULL DEFAULT '',
		trace_id VARCHAR(255) NOT NULL DEFAULT ''
    );`)
	if err != nil {
		panic(err)
	}

	// init job_summaries table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ` + jobSummaryModelName + ` (
		id VARCHAR(255) PRIMARY KEY NOT NULL DEFAULT '',
		success INTEGER NOT NULL DEFAULT 0,
		queueing INTEGER NOT NULL DEFAULT 0,
		retrying INTEGER NOT NULL DEFAULT 0,
		failure INTEGER NOT NULL DEFAULT 0,
		stopped INTEGER NOT NULL DEFAULT 0,
		is_loading BOOLEAN DEFAULT false
    );`)
	if err != nil {
		panic(err)
	}

	// init job_histories table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS task_queue_worker_job_histories (
		id INTEGER PRIMARY KEY NOT NULL,
		job_id VARCHAR(255),
		error_stack VARCHAR(255),
		status VARCHAR(255),
		error VARCHAR(255),
		trace_id VARCHAR(255),
		start_at VARCHAR(255),
		end_at VARCHAR(255)
    );`)
	if err != nil {
		panic(err)
	}

	indexList := map[string]struct {
		tableName, field string
	}{
		"idx_task_name":                   {jobModelName, "task_name"},
		"idx_status":                      {jobModelName, "status"},
		"idx_created_at":                  {jobModelName, "created_at"},
		"idx_args_err":                    {jobModelName, "arguments, error"},
		"idx_task_name_status":            {jobModelName, "task_name, status"},
		"idx_task_name_status_created_at": {jobModelName, "task_name, status, created_at"},
		"idx_task_name_summary":           {jobSummaryModelName, "id"},
		"idx_job_id_history":              {"task_queue_worker_job_histories", "job_id"},
		"idx_start_at_history":            {"task_queue_worker_job_histories", "start_at"},
	}
	for indexName, field := range indexList {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS ` + indexName + ` ON ` + field.tableName + ` (` + field.field + `)`)
		if err != nil {
			panic(err)
		}
	}

	sqlLitePersistent := &SQLPersistent{
		db:            db,
		queryReplacer: strings.NewReplacer("'", "''"),
	}
	sqlLitePersistent.summary = sqlLitePersistent
	return sqlLitePersistent
}

func (s *SQLPersistent) Ping(ctx context.Context) error {
	return s.db.Ping()
}
func (s *SQLPersistent) SetSummary(summary Summary) {
	s.summary = summary
}
func (s *SQLPersistent) Summary() Summary {
	return s.summary
}
func (s *SQLPersistent) FindAllJob(ctx context.Context, filter *Filter) (jobs []Job) {
	where, _ := s.toQueryFilter(filter)
	if filter.Sort == "" {
		filter.Sort = "-created_at"
	}
	sort := "ASC"
	if strings.HasPrefix(filter.Sort, "-") {
		sort = "DESC"
	}
	filter.Sort = strings.TrimPrefix(filter.Sort, "-")
	query := `SELECT id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id
		FROM ` + jobModelName + ` ` + where + ` ORDER BY ` + filter.Sort + ` ` + sort
	if !filter.ShowAll {
		query += fmt.Sprintf(` LIMIT %d OFFSET %d `, filter.Limit, (filter.Page-1)*filter.Limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		logger.LogE(err.Error())
		return jobs
	}
	defer rows.Close()

	for rows.Next() {
		var job Job
		var finishedAt sql.NullTime
		if err := rows.Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &job.CreatedAt,
			&finishedAt, &job.Status, &job.Error, &job.TraceID,
		); err != nil {
			return
		}
		job.FinishedAt = finishedAt.Time
		jobs = append(jobs, job)
	}

	return
}
func (s *SQLPersistent) FindJobByID(ctx context.Context, id string, excludeFields ...string) (job Job, err error) {
	var finishedAt sql.NullTime
	err = s.db.QueryRow(`SELECT id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id
			FROM `+jobModelName+` WHERE id='`+s.queryReplacer.Replace(id)+`'`).Scan(
		&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &job.CreatedAt,
		&finishedAt, &job.Status, &job.Error, &job.TraceID,
	)
	job.FinishedAt = finishedAt.Time
	if err != nil {
		logger.LogE(err.Error())
	}

	if len(excludeFields) == 0 {
		rows, err := s.db.Query(`SELECT error_stack, status, error, trace_id, start_at, end_at FROM task_queue_worker_job_histories
		WHERE job_id = '` + s.queryReplacer.Replace(id) + `' ORDER BY start_at DESC`)
		if err != nil {
			logger.LogE(err.Error())
			return job, err
		}
		defer rows.Close()

		for rows.Next() {
			var rh RetryHistory
			var startAt, endAt string
			rows.Scan(&rh.ErrorStack, &rh.Status, &rh.Error, &rh.TraceID, &startAt, &endAt)
			rh.StartAt, _ = time.Parse(time.RFC3339Nano, startAt)
			rh.EndAt, _ = time.Parse(time.RFC3339Nano, endAt)
			job.RetryHistories = append(job.RetryHistories, rh)
		}
	}

	return
}
func (s *SQLPersistent) CountAllJob(ctx context.Context, filter *Filter) (count int) {
	where, _ := s.toQueryFilter(filter)
	s.db.QueryRow(`SELECT COUNT(*) FROM ` + jobModelName + ` ` + where).Scan(&count)
	return
}
func (s *SQLPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary) {
	where, _ := s.toQueryFilter(filter)
	query := `SELECT COUNT(status), status, task_name FROM ` + jobModelName + ` ` + where + ` GROUP BY status, task_name`
	rows, err := s.db.Query(query)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	defer rows.Close()

	mapSummary := make(map[string]TaskSummary)
	for rows.Next() {
		var count int
		var status, taskName string
		rows.Scan(&count, &status, &taskName)
		summary := mapSummary[taskName]
		switch status {
		case string(statusSuccess):
			summary.Success += count
		case string(statusQueueing):
			summary.Queueing += count
		case string(statusRetrying):
			summary.Retrying += count
		case string(statusFailure):
			summary.Failure += count
		case string(statusStopped):
			summary.Stopped += count
		}
		mapSummary[taskName] = summary
	}

	for taskName, summary := range mapSummary {
		summary.TaskName = taskName
		summary.ID = taskName
		result = append(result, summary)
	}

	return
}
func (s *SQLPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) {
	var query string
	if job.ID == "" {
		job.ID = uuid.NewString()
		job.CreatedAt = time.Now()
		query = `INSERT INTO ` + jobModelName + ` (id, task_name, arguments, retries, max_retry, interval, created_at, 
			updated_at, finished_at, status, error, trace_id) VALUES (
				'` + s.queryReplacer.Replace(job.ID) + `',
				'` + s.queryReplacer.Replace(job.TaskName) + `',
				'` + s.queryReplacer.Replace(job.Arguments) + `',
				'` + candihelper.ToString(job.Retries) + `',
				'` + candihelper.ToString(job.MaxRetry) + `',
				'` + s.queryReplacer.Replace(job.Interval) + `',
				'` + job.CreatedAt.Format(time.RFC3339) + `',
				'` + time.Now().Format(time.RFC3339) + `',
				'` + job.FinishedAt.Format(time.RFC3339) + `',
				'` + s.queryReplacer.Replace(job.Status) + `',
				'` + s.queryReplacer.Replace(job.Error) + `',
				'` + s.queryReplacer.Replace(job.TraceID) + `'
			)`
	} else {
		query = `UPDATE ` + jobModelName + ` SET 
			task_name='` + s.queryReplacer.Replace(job.TaskName) + `', 
			arguments='` + s.queryReplacer.Replace(job.Arguments) + `', 
			retries='` + candihelper.ToString(job.Retries) + `', 
			max_retry='` + candihelper.ToString(job.MaxRetry) + `', 
			interval='` + s.queryReplacer.Replace(job.Interval) + `', 
			updated_at='` + time.Now().Format(time.RFC3339) + `', 
			finished_at='` + job.FinishedAt.Format(time.RFC3339) + `', 
			status='` + s.queryReplacer.Replace(job.Status) + `', 
			error='` + s.queryReplacer.Replace(job.Error) + `', 
			trace_id='` + s.queryReplacer.Replace(job.TraceID) + `'
			WHERE id = '` + s.queryReplacer.Replace(job.ID) + `'`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		logger.LogE(err.Error())
	}

	for _, rh := range retryHistories {
		_, err = s.db.Exec(`INSERT INTO task_queue_worker_job_histories (job_id, error_stack, status, error, trace_id, start_at, end_at) 
			VALUES (
				'` + s.queryReplacer.Replace(job.ID) + `',
				'` + s.queryReplacer.Replace(rh.ErrorStack) + `',
				'` + s.queryReplacer.Replace(rh.Status) + `',
				'` + s.queryReplacer.Replace(rh.Error) + `',
				'` + s.queryReplacer.Replace(rh.TraceID) + `',
				'` + rh.StartAt.Format(time.RFC3339Nano) + `',
				'` + rh.EndAt.Format(time.RFC3339Nano) + `'
			)`)
		if err != nil {
			logger.LogE(err.Error())
		}
	}
}
func (s *SQLPersistent) UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error) {
	where, err := s.toQueryFilter(filter)
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	tx.QueryRow(`SELECT COUNT(*) FROM ` + jobModelName + ` ` + where).Scan(&matchedCount)
	var setFields []string
	for field, value := range updated {
		setFields = append(setFields, field+"='"+s.queryReplacer.Replace(candihelper.ToString(value))+"'")
	}
	res, err := tx.Exec(`UPDATE ` + jobModelName + ` SET ` + strings.Join(setFields, ",") + ` ` + where)
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	affectedRow, _ = res.RowsAffected()

	if filter.JobID != nil {
		for _, rh := range retryHistories {
			_, err = tx.Exec(`INSERT INTO task_queue_worker_job_histories (job_id, error_stack, status, error, trace_id, start_at, end_at) 
			VALUES (
				'` + s.queryReplacer.Replace(*filter.JobID) + `',
				'` + s.queryReplacer.Replace(rh.ErrorStack) + `',
				'` + s.queryReplacer.Replace(rh.Status) + `',
				'` + s.queryReplacer.Replace(rh.Error) + `',
				'` + s.queryReplacer.Replace(rh.TraceID) + `',
				'` + rh.StartAt.Format(time.RFC3339Nano) + `',
				'` + rh.EndAt.Format(time.RFC3339Nano) + `'
			)`)
			if err != nil {
				logger.LogE(err.Error())
			}
		}
	}
	return
}
func (s *SQLPersistent) CleanJob(ctx context.Context, filter *Filter) (affectedRow int64) {
	where, err := s.toQueryFilter(filter)
	if err != nil {
		logger.LogE(err.Error())
		return affectedRow
	}

	res, err := s.db.Exec(`DELETE FROM ` + jobModelName + ` ` + where)
	if err != nil {
		logger.LogE(err.Error())
		return affectedRow
	}
	affectedRow, _ = res.RowsAffected()
	return
}
func (s *SQLPersistent) DeleteJob(ctx context.Context, id string) (job Job, err error) {
	err = s.db.QueryRow(`SELECT id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id
		FROM `+jobModelName+` WHERE id='`+s.queryReplacer.Replace(id)+`'`).Scan(
		&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &job.CreatedAt,
		&job.FinishedAt, &job.Status, &job.Error, &job.TraceID,
	)
	_, err = s.db.Exec(`DELETE FROM ` + jobModelName + ` WHERE id='` + s.queryReplacer.Replace(id) + `'`)
	if err != nil {
		logger.LogE(err.Error())
	}
	return
}

// summary
func (s *SQLPersistent) FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary) {
	var where string
	if filter.TaskName != "" {
		where = ` WHERE id = '` + s.queryReplacer.Replace(filter.TaskName) + `'`
	} else if len(filter.TaskNameList) > 0 {
		var taskNameList []string
		for _, taskName := range filter.TaskNameList {
			taskNameList = append(taskNameList, "'"+s.queryReplacer.Replace(taskName)+"'")
		}
		where = " WHERE id IN (" + strings.Join(taskNameList, ",") + ")"
	}
	query := `SELECT id, success, queueing, retrying, failure, stopped, is_loading FROM ` + jobSummaryModelName + where
	rows, err := s.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var detail TaskSummary
		rows.Scan(&detail.TaskName, &detail.Success, &detail.Queueing, &detail.Retrying,
			&detail.Failure, &detail.Stopped, &detail.IsLoading)
		detail.ID = detail.TaskName
		result = append(result, detail)
	}

	if len(filter.Statuses) > 0 {
		for i, res := range result {
			mapRes := res.ToMapResult()
			newCount := map[string]int{}
			for _, status := range filter.Statuses {
				newCount[strings.ToUpper(status)] = mapRes[strings.ToUpper(status)]
			}
			res.SetValue(newCount)
			result[i] = res
		}
	}

	return
}
func (s *SQLPersistent) FindDetailSummary(ctx context.Context, taskName string) (result TaskSummary) {
	err := s.db.QueryRow(`SELECT id, success, queueing, retrying, failure, stopped, is_loading
		FROM `+jobSummaryModelName+` WHERE id='`+s.queryReplacer.Replace(taskName)+`'`).
		Scan(&result.TaskName, &result.Success, &result.Queueing, &result.Retrying,
			&result.Failure, &result.Stopped, &result.IsLoading)
	if err != nil {
		logger.LogE(err.Error())
	}
	result.ID = result.TaskName
	return
}
func (s *SQLPersistent) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {
	var setFields []string
	for field, value := range updated {
		if field == "" {
			continue
		}
		field = strings.ToLower(field)
		setFields = append(setFields, field+"='"+candihelper.ToString(value)+"'")
	}
	query := `UPDATE ` + jobSummaryModelName + ` SET ` + strings.Join(setFields, ",") + ` WHERE id='` + s.queryReplacer.Replace(taskName) + `'`
	res, err := s.db.Exec(query)
	if err != nil {
		logger.LogE(err.Error())
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		query := fmt.Sprintf(`INSERT INTO %s (id, success, queueing, retrying, failure, stopped) VALUES ('%s', '%d', '%d', '%d', '%d', '%d')`,
			jobSummaryModelName, s.queryReplacer.Replace(taskName), candihelper.ToInt(updated["success"]), candihelper.ToInt(updated["queueing"]),
			candihelper.ToInt(updated["retrying"]), candihelper.ToInt(updated["failure"]), candihelper.ToInt(updated["stopped"]))
		_, err := s.db.Exec(query)
		if err != nil {
			logger.LogE(err.Error())
		}
	}
	return
}
func (s *SQLPersistent) IncrementSummary(ctx context.Context, taskName string, incr map[string]int64) {
	var setFields []string
	for field, value := range incr {
		if field == "" {
			continue
		}
		field = strings.ToLower(field)
		val := candihelper.ToString(value)
		if value >= 0 {
			val = "+" + candihelper.ToString(value)
		}
		setFields = append(setFields, field+"="+field+val)
	}
	query := `UPDATE ` + jobSummaryModelName + ` SET ` + strings.Join(setFields, ",") + ` WHERE id='` + s.queryReplacer.Replace(taskName) + `'`
	res, err := s.db.Exec(query)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		query := fmt.Sprintf(`INSERT INTO %s (id, success, queueing, retrying, failure, stopped) VALUES ('%s', '%d', '%d', '%d', '%d', '%d')`,
			jobSummaryModelName, s.queryReplacer.Replace(taskName), candihelper.ToInt(incr["success"]), candihelper.ToInt(incr["queueing"]),
			candihelper.ToInt(incr["retrying"]), candihelper.ToInt(incr["failure"]), candihelper.ToInt(incr["stopped"]))
		_, err := s.db.Exec(query)
		if err != nil {
			logger.LogE(err.Error())
		}
	}
	return
}

func (s *SQLPersistent) toQueryFilter(f *Filter) (where string, err error) {

	var conditions []string
	if f.TaskName != "" {
		conditions = append(conditions, "task_name='"+s.queryReplacer.Replace(f.TaskName)+"'")
	} else if len(f.TaskNameList) > 0 {
		var taskNameList []string
		for _, taskName := range f.TaskNameList {
			taskNameList = append(taskNameList, "'"+s.queryReplacer.Replace(taskName)+"'")
		}
		conditions = append(conditions, "task_name IN ("+strings.Join(taskNameList, ",")+")")
	}

	if f.JobID != nil && *f.JobID != "" {
		conditions = append(conditions, "id='"+s.queryReplacer.Replace(*f.JobID)+"'")
	}
	if f.Search != nil && *f.Search != "" {
		conditions = append(conditions, `(arguments LIKE '%%`+*f.Search+`%%' OR error LIKE '%%`+*f.Search+`%%')`)
	}
	if len(f.Statuses) > 0 {
		var statuses []string
		for _, status := range f.Statuses {
			statuses = append(statuses, "'"+s.queryReplacer.Replace(status)+"'")
		}
		conditions = append(conditions, "status IN ("+strings.Join(statuses, ",")+")")
	}
	if f.Status != nil {
		conditions = append(conditions, "status='"+s.queryReplacer.Replace(*f.Status)+"'")
	}
	if !f.StartDate.IsZero() && !f.EndDate.IsZero() {
		conditions = append(conditions, "created_at BETWEEN '"+f.StartDate.Format(time.RFC3339)+"' AND '"+f.EndDate.Format(time.RFC3339)+"'")
	}

	if len(conditions) == 0 {
		return where, errors.New("empty filter")
	}

	where = " WHERE " + strings.Join(conditions, " AND ")
	return
}
