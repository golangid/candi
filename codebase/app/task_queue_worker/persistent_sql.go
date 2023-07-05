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
		driverName    string
		summary       Summary
		queryReplacer *strings.Replacer
	}
)

// NewSQLPersistent init new persistent SQL
func NewSQLPersistent(db *sql.DB) *SQLPersistent {
	sqlPersistent := &SQLPersistent{
		db:            db,
		queryReplacer: strings.NewReplacer("'", "''"),
	}

	dbDriverType := fmt.Sprintf("%T", db.Driver())
	driverName, ok := map[string]string{
		"*pq.Driver":            "postgres",
		"*mysql.MySQLDriver":    "mysql",
		"*sqlite3.SQLiteDriver": "sqlite3",
	}[dbDriverType]
	if !ok {
		panic("Unknown SQL persistent driver " + dbDriverType + " for Task Queue Worker. Only support postgres, mysql, or sqlite3 driver")
	}
	sqlPersistent.driverName = driverName
	sqlPersistent.summary = sqlPersistent
	sqlPersistent.initTable(db)
	return sqlPersistent
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
	query := "SELECT " +
		s.formatColumnName("id", "task_name", "arguments", "retries", "max_retry", "interval", "created_at", "finished_at", "status", "error", "result", "trace_id", "current_progress", "max_progress") +
		" FROM " + jobModelName + " " + where + " ORDER BY " + s.formatColumnName(strings.TrimPrefix(filter.Sort, "-")) + " " + sort
	if !filter.ShowAll {
		query += fmt.Sprintf(` LIMIT %d OFFSET %d `, filter.Limit, filter.CalculateOffset())
	}
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		logger.LogE(err.Error())
		return jobs
	}
	defer rows.Close()

	for rows.Next() {
		var job Job
		var createdAt string
		var finishedAt, result sql.NullString
		if err := rows.Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &createdAt,
			&finishedAt, &job.Status, &job.Error, &result, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
		); err != nil {
			logger.LogE(err.Error())
			return
		}
		job.CreatedAt = s.parseDateString(createdAt).Time
		job.FinishedAt = s.parseDateString(finishedAt.String).Time
		job.Result = result.String
		jobs = append(jobs, job)
	}

	return
}
func (s *SQLPersistent) FindJobByID(ctx context.Context, id string, filterHistory *Filter) (job Job, err error) {
	var finishedAt, result sql.NullString
	var createdAt string
	err = s.db.QueryRowContext(ctx, `SELECT `+
		s.formatColumnName("id", "task_name", "arguments", "retries", "max_retry", "interval", "created_at", "finished_at", "status", "error", "result", "trace_id", "current_progress", "max_progress")+
		` FROM `+jobModelName+` WHERE id='`+id+`'`).
		Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &createdAt,
			&finishedAt, &job.Status, &job.Error, &result, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
		)
	if err != nil {
		logger.LogE(err.Error())
		return job, err
	}
	job.CreatedAt = s.parseDateString(createdAt).Time
	job.FinishedAt = s.parseDateString(finishedAt.String).Time
	job.Result = result.String

	if filterHistory != nil {
		query := `SELECT ` + s.formatColumnName("error_stack", "status", "error", "result", "trace_id", "start_at", "end_at") +
			` FROM task_queue_worker_job_histories WHERE job_id = '` + id + `' ORDER BY start_at DESC `
		rows, err := s.db.Query(query + fmt.Sprintf(` LIMIT %d OFFSET %d`, filterHistory.Limit, filterHistory.CalculateOffset()))
		if err != nil {
			logger.LogE(err.Error())
			return job, err
		}
		defer rows.Close()

		for rows.Next() {
			var rh RetryHistory
			var startAt, endAt string
			var result sql.NullString
			rows.Scan(&rh.ErrorStack, &rh.Status, &rh.Error, &result, &rh.TraceID, &startAt, &endAt)
			rh.StartAt = s.parseDateString(startAt).Time
			rh.EndAt = s.parseDateString(endAt).Time
			rh.Result = result.String
			job.RetryHistories = append(job.RetryHistories, rh)
		}
		s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM task_queue_worker_job_histories WHERE job_id = '`+id+`'`).Scan(&filterHistory.Count)
	}

	return
}
func (s *SQLPersistent) CountAllJob(ctx context.Context, filter *Filter) (count int) {
	where, _ := s.toQueryFilter(filter)
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+jobModelName+` `+where).Scan(&count); err != nil {
		logger.LogE(err.Error())
	}
	return
}
func (s *SQLPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary) {
	where, _ := s.toQueryFilter(filter)
	query := "SELECT COUNT(" + s.formatColumnName("status") + "), " + s.formatColumnName("status", "task_name") +
		" FROM " + jobModelName + " " + where +
		" GROUP BY " + s.formatColumnName("status", "task_name")
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
		case string(StatusSuccess):
			summary.Success += count
		case string(StatusQueueing):
			summary.Queueing += count
		case string(StatusRetrying):
			summary.Retrying += count
		case string(StatusFailure):
			summary.Failure += count
		case string(StatusStopped):
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
func (s *SQLPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) (err error) {
	var query string
	var args []interface{}
	if job.ID == "" {
		job.ID = uuid.NewString()
		job.CreatedAt = time.Now()
		args = []interface{}{
			job.ID, job.TaskName, job.Arguments, job.Retries, job.MaxRetry, job.Interval, s.parseDate(job.CreatedAt), s.parseDate(time.Now()), s.parseDate(job.FinishedAt),
			job.Status, job.Error, job.Result, job.TraceID, job.CurrentProgress, job.MaxProgress,
		}
		query = "INSERT INTO " + jobModelName + " (" +
			s.formatColumnName("id", "task_name", "arguments", "retries", "max_retry", "interval", "created_at", "updated_at", "finished_at", "status", "error", "result", "trace_id", "current_progress", "max_progress") +
			") VALUES (" + s.parameterize(len(args)) + ")"
	} else {
		args = []interface{}{
			job.TaskName, job.Arguments, job.Retries, job.MaxRetry, job.Interval, s.parseDate(time.Now()), s.parseDate(job.FinishedAt), job.Status,
			job.Error, job.Result, job.TraceID, job.CurrentProgress, job.MaxProgress,
		}
		query = `UPDATE ` + jobModelName + ` SET ` +
			s.parameterizeForUpdate("task_name", "arguments", "retries", "max_retry", "interval", "updated_at", "finished_at", "status", "error", "result", "trace_id", "current_progress", "max_progress") +
			` WHERE id = '` + job.ID + `'`
	}
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}

	for _, rh := range retryHistories {
		args := []interface{}{job.ID, rh.ErrorStack, rh.Status, rh.Error, rh.Result, rh.TraceID, rh.StartAt, rh.EndAt}
		_, err = s.db.ExecContext(ctx, `INSERT INTO task_queue_worker_job_histories (`+
			s.formatColumnName("job_id", "error_stack", "status", "error", "result", "trace_id", "start_at", "end_at")+
			`) VALUES (`+s.parameterize(len(args))+`)`, args...)
		if err != nil {
			logger.LogE(err.Error())
			return err
		}
	}

	return nil
}
func (s *SQLPersistent) UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error) {
	where, err := s.toQueryFilter(filter)
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+jobModelName+` `+where).Scan(&matchedCount)

	var columns []string
	var args []interface{}
	updated["updated_at"] = time.Now()
	for field, value := range updated {
		if v, ok := value.(time.Time); ok {
			value = s.parseDate(v)
		}
		columns = append(columns, field)
		args = append(args, value)
	}
	query := `UPDATE ` + jobModelName + ` SET ` + s.parameterizeForUpdate(columns...) + ` ` + where

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	affectedRow, _ = res.RowsAffected()

	if filter.JobID != nil {
		for _, rh := range retryHistories {
			args := []interface{}{filter.JobID, rh.ErrorStack, rh.Status, rh.Error, rh.Result, rh.TraceID, s.parseDate(rh.StartAt), s.parseDate(rh.EndAt)}
			_, err = s.db.ExecContext(ctx, `INSERT INTO task_queue_worker_job_histories (`+
				s.formatColumnName("job_id", "error_stack", "status", "error", "result", "trace_id", "start_at", "end_at")+
				`) VALUES (`+s.parameterize(len(args))+`)`, args...)
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

	_, err = s.db.Exec(`DELETE FROM task_queue_worker_job_histories WHERE job_id IN (SELECT id FROM ` + jobModelName + ` ` + where + `)`)
	if err != nil {
		logger.LogE(err.Error())
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
	var createdAt, finishedAt, result sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT `+
		s.formatColumnName("id", "task_name", "arguments", "retries", "max_retry", "interval", "created_at", "finished_at", "status", "error", "result", "trace_id", "current_progress", "max_progress")+
		` FROM `+jobModelName+` WHERE id=`+s.parameterize(1), id).
		Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &createdAt,
			&finishedAt, &job.Status, &job.Error, &result, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
		)
	job.CreatedAt = s.parseDateString(createdAt.String).Time
	job.FinishedAt = s.parseDateString(finishedAt.String).Time
	job.Result = result.String
	logger.LogIfError(err)
	_, err = s.db.Exec(`DELETE FROM ` + jobModelName + ` WHERE id='` + id + `'`)
	logger.LogIfError(err)
	_, err = s.db.Exec(`DELETE FROM task_queue_worker_job_histories WHERE job_id='` + id + `'`)
	logger.LogIfError(err)
	return
}

// summary
func (s *SQLPersistent) FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary) {
	var where string
	var args []interface{}
	if filter.TaskName != "" {
		args = append(args, filter.TaskName)
		where = ` WHERE id = ` + s.parameterize(len(args))
	} else if len(filter.TaskNameList) > 0 {
		for _, taskName := range filter.TaskNameList {
			args = append(args, taskName)
		}
		where = " WHERE id IN (" + s.parameterize(len(args)) + ")"
	}
	query := `SELECT ` + s.formatColumnName("id", "success", "queueing", "retrying", "failure", "stopped", "is_loading") +
		` FROM ` + jobSummaryModelName + where + " ORDER BY id ASC"
	rows, err := s.db.Query(query, args...)
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
	s.db.QueryRowContext(ctx, `SELECT `+s.formatColumnName("id", "success", "queueing", "retrying", "failure", "stopped", "is_loading")+
		` FROM `+jobSummaryModelName+` WHERE id=`+s.parameterize(1), taskName).
		Scan(&result.TaskName, &result.Success, &result.Queueing, &result.Retrying,
			&result.Failure, &result.Stopped, &result.IsLoading)
	result.ID = result.TaskName
	return
}
func (s *SQLPersistent) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {
	var columns []string
	var args []interface{}
	for field, value := range updated {
		if field == "" {
			continue
		}
		columns = append(columns, field)
		args = append(args, value)
	}
	query := `UPDATE ` + jobSummaryModelName + ` SET ` + s.parameterizeForUpdate(columns...) + ` WHERE id='` + taskName + `'`
	res, err := s.db.Exec(query, args...)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		args := []interface{}{
			taskName, candihelper.ToInt(updated["success"]), candihelper.ToInt(updated["queueing"]), candihelper.ToInt(updated["retrying"]),
			candihelper.ToInt(updated["failure"]), candihelper.ToInt(updated["stopped"]), updated["is_loading"],
		}
		query := `INSERT INTO ` + jobSummaryModelName + ` (` +
			s.formatColumnName("id", "success", "queueing", "retrying", "failure", "stopped", "is_loading") +
			`) VALUES (` + s.parameterize(len(args)) + `)`
		_, err := s.db.Exec(query, args...)
		if err != nil {
			logger.LogE(err.Error())
		}
	}
	return
}
func (s *SQLPersistent) IncrementSummary(ctx context.Context, taskName string, incr map[string]int64) {
	if len(incr) == 0 {
		return
	}

	var setFields []string
	for field, value := range incr {
		if field == "" {
			continue
		}
		field = s.formatColumnName(strings.ToLower(field))
		val := candihelper.ToString(value)
		if value >= 0 {
			val = "+" + candihelper.ToString(value)
		}
		setFields = append(setFields, field+"="+field+val)
	}
	query := `UPDATE ` + jobSummaryModelName + ` SET ` + strings.Join(setFields, ",") + ` WHERE id='` + taskName + `'`
	res, err := s.db.Exec(query)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		args := []interface{}{
			taskName, candihelper.ToInt(incr["success"]), candihelper.ToInt(incr["queueing"]), candihelper.ToInt(incr["retrying"]),
			candihelper.ToInt(incr["failure"]), candihelper.ToInt(incr["stopped"]),
		}
		query := `INSERT INTO ` + jobSummaryModelName + ` (` +
			s.formatColumnName("id", "success", "queueing", "retrying", "failure", "stopped") +
			`) VALUES (` + s.parameterize(len(args)) + `)`
		_, err := s.db.Exec(query, args...)
		if err != nil {
			logger.LogE(err.Error())
		}
	}
	return
}
func (s *SQLPersistent) DeleteAllSummary(ctx context.Context, filter *Filter) {
	var where string
	if len(filter.ExcludeTaskNameList) > 0 {
		where = "WHERE id NOT IN " + s.toMultiParamQuery(filter.ExcludeTaskNameList)
	}
	_, err := s.db.Exec(`DELETE FROM ` + jobSummaryModelName + ` ` + where)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
}
func (s *SQLPersistent) Type() string {
	var versionFunc, version string
	switch s.driverName {
	case "postgres", "mysql":
		versionFunc = "version()"
	case "sqlite3":
		versionFunc = "sqlite_version()"
	default:
		return "SQL Persistent"
	}
	s.db.QueryRow(`SELECT ` + versionFunc).Scan(&version)
	return "SQL Persistent (driver: " + s.driverName + ") " + version
}

func (s *SQLPersistent) toQueryFilter(f *Filter) (where string, err error) {
	var conditions []string
	if f.TaskName != "" {
		conditions = append(conditions, s.formatColumnName("task_name")+"='"+s.queryReplacer.Replace(f.TaskName)+"'")
	} else if len(f.TaskNameList) > 0 {
		conditions = append(conditions, s.formatColumnName("task_name")+" IN "+s.toMultiParamQuery(f.TaskNameList))
	} else if len(f.ExcludeTaskNameList) > 0 {
		conditions = append(conditions, s.formatColumnName("task_name")+" NOT IN "+s.toMultiParamQuery(f.ExcludeTaskNameList))
	}

	if f.JobID != nil && *f.JobID != "" {
		conditions = append(conditions, "id='"+s.queryReplacer.Replace(*f.JobID)+"'")
	}
	if f.Search != nil && *f.Search != "" {
		conditions = append(conditions, "("+s.formatColumnName("arguments")+" LIKE '%%"+*f.Search+"%%' OR "+
			s.formatColumnName("error")+" LIKE '%%"+*f.Search+"%%')")
	}
	if len(f.Statuses) > 0 {
		conditions = append(conditions, s.formatColumnName("status")+" IN "+s.toMultiParamQuery(f.Statuses))
	}
	if f.Status != nil {
		conditions = append(conditions, s.formatColumnName("status")+"='"+s.queryReplacer.Replace(*f.Status)+"'")
	}
	if startDate, endDate := f.ParseStartEndDate(); !startDate.IsZero() && !endDate.IsZero() {
		conditions = append(conditions, s.formatColumnName("created_at")+" BETWEEN '"+startDate.Format(time.RFC3339)+"' AND '"+endDate.Format(time.RFC3339)+"'")
	}
	if f.BeforeCreatedAt != nil && !f.BeforeCreatedAt.IsZero() {
		conditions = append(conditions, s.formatColumnName("created_at")+" <= '"+f.BeforeCreatedAt.Format(time.RFC3339)+"'")
	}

	if len(conditions) == 0 {
		return where, errors.New("empty filter")
	}

	where = " WHERE " + strings.Join(conditions, " AND ")
	return
}

func (s *SQLPersistent) GetAllConfiguration(ctx context.Context) (cfg []Configuration, err error) {
	query := "SELECT " + s.formatColumnName("key", "name", "value", "is_active") + " FROM " + configurationModelName +
		" ORDER BY " + s.formatColumnName("key")
	rows, err := s.db.Query(query)
	if err != nil {
		logger.LogE(err.Error())
		return cfg, err
	}
	defer rows.Close()
	for rows.Next() {
		var config Configuration
		rows.Scan(&config.Key, &config.Name, &config.Value, &config.IsActive)
		cfg = append(cfg, config)
	}
	return
}

func (s *SQLPersistent) GetConfiguration(key string) (cfg Configuration, err error) {
	query := "SELECT " + s.formatColumnName("key", "name", "value", "is_active") + " FROM " + configurationModelName +
		` WHERE ` + s.parameterizeByColumnAndNumber("key", 1)
	err = s.db.QueryRow(query, key).Scan(&cfg.Key, &cfg.Name, &cfg.Value, &cfg.IsActive)
	return
}

func (s *SQLPersistent) SetConfiguration(cfg *Configuration) (err error) {
	args := []interface{}{cfg.Name, cfg.Value, cfg.IsActive}
	query := `UPDATE ` + configurationModelName + ` SET ` + s.parameterizeForUpdate("name", "value", "is_active") +
		` WHERE ` + s.parameterizeByColumnAndNumber("key", len(args)+1)
	args = append(args, cfg.Key)
	res, err := s.db.Exec(query, args...)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		args := []interface{}{cfg.Key, cfg.Name, cfg.Value, cfg.IsActive}
		query := `INSERT INTO ` + configurationModelName +
			` (` + s.formatColumnName("key", "name", "value", "is_active") + ` ) VALUES (` + s.parameterize(len(args)) + `)`
		s.db.Exec(query, args...)
	}
	return nil
}
