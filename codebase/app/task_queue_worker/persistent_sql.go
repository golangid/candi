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
	filter.Sort = s.formatColumnName(strings.TrimPrefix(filter.Sort, "-"))
	query := "SELECT " +
		s.formatMultiColumnName("id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id, current_progress, max_progress") +
		" FROM " + jobModelName + " " + where + " ORDER BY " + filter.Sort + " " + sort
	if !filter.ShowAll {
		query += fmt.Sprintf(` LIMIT %d OFFSET %d `, filter.Limit, filter.CalculateOffset())
	}
	rows, err := s.db.Query(query)
	if err != nil {
		logger.LogE(err.Error())
		return jobs
	}
	defer rows.Close()

	for rows.Next() {
		var job Job
		var createdAt string
		var finishedAt sql.NullString
		if err := rows.Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &createdAt,
			&finishedAt, &job.Status, &job.Error, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
		); err != nil {
			logger.LogE(err.Error())
			return
		}
		job.CreatedAt = s.parseDateString(createdAt)
		job.FinishedAt = s.parseDateString(finishedAt.String)
		jobs = append(jobs, job)
	}

	return
}
func (s *SQLPersistent) FindJobByID(ctx context.Context, id string, filterHistory *Filter) (job Job, err error) {
	var finishedAt sql.NullString
	var createdAt string
	err = s.db.QueryRow(`SELECT `+
		s.formatMultiColumnName("id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id, current_progress, max_progress")+
		` FROM `+jobModelName+` WHERE id='`+s.queryReplacer.Replace(id)+`'`).
		Scan(
			&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &createdAt,
			&finishedAt, &job.Status, &job.Error, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
		)
	if err != nil {
		logger.LogE(err.Error())
		return job, err
	}
	job.CreatedAt = s.parseDateString(createdAt)
	job.FinishedAt = s.parseDateString(finishedAt.String)

	if filterHistory != nil {
		query := `SELECT error_stack, status, error, trace_id, start_at, end_at FROM task_queue_worker_job_histories
			WHERE job_id = '` + s.queryReplacer.Replace(id) + `' ORDER BY start_at DESC `
		rows, err := s.db.Query(query + fmt.Sprintf(` LIMIT %d OFFSET %d`, filterHistory.Limit, filterHistory.CalculateOffset()))
		if err != nil {
			logger.LogE(err.Error())
			return job, err
		}
		defer rows.Close()

		for rows.Next() {
			var rh RetryHistory
			var startAt, endAt string
			rows.Scan(&rh.ErrorStack, &rh.Status, &rh.Error, &rh.TraceID, &startAt, &endAt)
			rh.StartAt = s.parseDateString(startAt)
			rh.EndAt = s.parseDateString(endAt)
			job.RetryHistories = append(job.RetryHistories, rh)
		}
		s.db.QueryRow(`SELECT COUNT(*) FROM task_queue_worker_job_histories WHERE job_id = '` + s.queryReplacer.Replace(id) + `'`).
			Scan(&filterHistory.Count)
	}

	return
}
func (s *SQLPersistent) CountAllJob(ctx context.Context, filter *Filter) (count int) {
	where, _ := s.toQueryFilter(filter)
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM ` + jobModelName + ` ` + where).Scan(&count); err != nil {
		logger.LogE(err.Error())
	}
	return
}
func (s *SQLPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary) {
	where, _ := s.toQueryFilter(filter)
	query := `SELECT COUNT(` + s.formatColumnName("status") + `), ` + s.formatMultiColumnName("status, task_name") +
		` FROM ` + jobModelName + ` ` + where +
		` GROUP BY` + s.formatMultiColumnName("status, task_name")
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
	finishedAt := s.parseDate(job.FinishedAt)
	if finishedAt == "" {
		finishedAt = "NULL"
	}
	if job.ID == "" {
		job.ID = uuid.NewString()
		job.CreatedAt = time.Now()
		query = `INSERT INTO ` + jobModelName + ` (` +
			s.formatMultiColumnName("id, task_name, arguments, retries, max_retry, interval, created_at, updated_at, finished_at, status, error, trace_id, current_progress, max_progress") +
			`) VALUES (
				'` + s.queryReplacer.Replace(job.ID) + `',
				'` + s.queryReplacer.Replace(job.TaskName) + `',
				'` + s.queryReplacer.Replace(job.Arguments) + `',
				'` + candihelper.ToString(job.Retries) + `',
				'` + candihelper.ToString(job.MaxRetry) + `',
				'` + s.queryReplacer.Replace(job.Interval) + `',
				'` + s.parseDate(job.CreatedAt) + `',
				'` + s.parseDate(time.Now()) + `',
				` + finishedAt + `,
				'` + s.queryReplacer.Replace(job.Status) + `',
				'` + s.queryReplacer.Replace(job.Error) + `',
				'` + s.queryReplacer.Replace(job.TraceID) + `',
				'` + candihelper.ToString(job.CurrentProgress) + `',
				'` + candihelper.ToString(job.MaxProgress) + `'
			)`
	} else {
		query = `UPDATE ` + jobModelName + ` SET 
			` + s.formatColumnName("task_name") + `='` + s.queryReplacer.Replace(job.TaskName) + `', 
			` + s.formatColumnName("arguments") + `='` + s.queryReplacer.Replace(job.Arguments) + `', 
			` + s.formatColumnName("retries") + `='` + candihelper.ToString(job.Retries) + `', 
			` + s.formatColumnName("max_retry") + `='` + candihelper.ToString(job.MaxRetry) + `', 
			` + s.formatColumnName("interval") + `='` + s.queryReplacer.Replace(job.Interval) + `', 
			` + s.formatColumnName("updated_at") + `='` + s.parseDate(time.Now()) + `', 
			` + s.formatColumnName("finished_at") + `=` + finishedAt + `, 
			` + s.formatColumnName("status") + `='` + s.queryReplacer.Replace(job.Status) + `', 
			` + s.formatColumnName("error") + `='` + s.queryReplacer.Replace(job.Error) + `', 
			` + s.formatColumnName("trace_id") + `='` + s.queryReplacer.Replace(job.TraceID) + `', 
			` + s.formatColumnName("current_progress") + `='` + candihelper.ToString(job.CurrentProgress) + `', 
			` + s.formatColumnName("max_progress") + `='` + candihelper.ToString(job.MaxProgress) + `'
			WHERE id = '` + s.queryReplacer.Replace(job.ID) + `'`
	}

	_, err = s.db.ExecContext(ctx, query)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}

	for _, rh := range retryHistories {
		_, err = s.db.ExecContext(ctx, `INSERT INTO task_queue_worker_job_histories (`+
			s.formatMultiColumnName("job_id, error_stack, status, error, trace_id, start_at, end_at")+
			`) VALUES (
				'`+s.queryReplacer.Replace(job.ID)+`',
				'`+s.queryReplacer.Replace(rh.ErrorStack)+`',
				'`+s.queryReplacer.Replace(rh.Status)+`',
				'`+s.queryReplacer.Replace(rh.Error)+`',
				'`+s.queryReplacer.Replace(rh.TraceID)+`',
				'`+s.parseDate(rh.StartAt)+`',
				'`+s.parseDate(rh.EndAt)+`'
			)`)
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
	var setFields []string
	updated["updated_at"] = time.Now()
	for field, value := range updated {
		if t, ok := value.(time.Time); ok {
			value = s.parseDate(t)
		}
		field = s.formatColumnName(field)
		val := s.queryReplacer.Replace(candihelper.ToString(value))
		switch value.(type) {
		case bool:
			setFields = append(setFields, field+"="+val)
		default:
			setFields = append(setFields, field+"='"+val+"'")
		}
	}
	query := `UPDATE ` + jobModelName + ` SET ` + strings.Join(setFields, ",") + ` ` + where
	res, err := s.db.ExecContext(ctx, query)
	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}
	affectedRow, _ = res.RowsAffected()

	if filter.JobID != nil {
		for _, rh := range retryHistories {
			_, err = s.db.ExecContext(ctx, `INSERT INTO task_queue_worker_job_histories (`+
				s.formatMultiColumnName("job_id, error_stack, status, error, trace_id, start_at, end_at")+
				`) VALUES (
				'`+s.queryReplacer.Replace(*filter.JobID)+`',
				'`+s.queryReplacer.Replace(rh.ErrorStack)+`',
				'`+s.queryReplacer.Replace(rh.Status)+`',
				'`+s.queryReplacer.Replace(rh.Error)+`',
				'`+s.queryReplacer.Replace(rh.TraceID)+`',
				'`+s.parseDate(rh.StartAt)+`',
				'`+s.parseDate(rh.EndAt)+`'
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
	err = s.db.QueryRow(`SELECT `+
		s.formatMultiColumnName(`id, task_name, arguments, retries, max_retry, interval, created_at, finished_at, status, error, trace_id, current_progress, max_progress`)+
		` FROM `+jobModelName+` WHERE id='`+s.queryReplacer.Replace(id)+`'`).Scan(
		&job.ID, &job.TaskName, &job.Arguments, &job.Retries, &job.MaxRetry, &job.Interval, &job.CreatedAt,
		&job.FinishedAt, &job.Status, &job.Error, &job.TraceID, &job.CurrentProgress, &job.MaxProgress,
	)
	_, err = s.db.Exec(`DELETE FROM ` + jobModelName + ` WHERE id='` + s.queryReplacer.Replace(id) + `'`)
	if err != nil {
		logger.LogE(err.Error())
	}
	_, err = s.db.Exec(`DELETE FROM task_queue_worker_job_histories WHERE job_id='` + s.queryReplacer.Replace(id) + `'`)
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
	query := `SELECT ` + s.formatMultiColumnName("id, success, queueing, retrying, failure, stopped, is_loading") +
		` FROM ` + jobSummaryModelName + where + " ORDER BY id ASC"
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
	s.db.QueryRow(`SELECT `+s.formatMultiColumnName("id, success, queueing, retrying, failure, stopped, is_loading")+
		` FROM `+jobSummaryModelName+` WHERE id='`+s.queryReplacer.Replace(taskName)+`'`).
		Scan(&result.TaskName, &result.Success, &result.Queueing, &result.Retrying,
			&result.Failure, &result.Stopped, &result.IsLoading)
	result.ID = result.TaskName
	return
}
func (s *SQLPersistent) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {
	var setFields []string
	for field, value := range updated {
		if field == "" {
			continue
		}
		field = s.formatColumnName(strings.ToLower(field))
		switch value.(type) {
		case bool:
			setFields = append(setFields, field+"="+candihelper.ToString(value))
		default:
			setFields = append(setFields, field+"='"+candihelper.ToString(value)+"'")
		}
	}
	query := `UPDATE ` + jobSummaryModelName + ` SET ` + strings.Join(setFields, ",") + ` WHERE id='` + s.queryReplacer.Replace(taskName) + `'`
	res, err := s.db.Exec(query)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		query := fmt.Sprintf(`INSERT INTO %s (`+
			s.formatMultiColumnName("id, success, queueing, retrying, failure, stopped, is_loading")+
			`) VALUES ('%s', '%d', '%d', '%d', '%d', '%d', %s)`,
			jobSummaryModelName, s.queryReplacer.Replace(taskName), candihelper.ToInt(updated["success"]), candihelper.ToInt(updated["queueing"]),
			candihelper.ToInt(updated["retrying"]), candihelper.ToInt(updated["failure"]), candihelper.ToInt(updated["stopped"]),
			s.queryReplacer.Replace(candihelper.ToString(updated["is_loading"])))
		_, err := s.db.Exec(query)
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
		query := fmt.Sprintf(`INSERT INTO %s (`+
			s.formatMultiColumnName("id, success, queueing, retrying, failure, stopped")+
			`) VALUES ('%s', '%d', '%d', '%d', '%d', '%d')`,
			jobSummaryModelName, s.queryReplacer.Replace(taskName), candihelper.ToInt(incr["success"]), candihelper.ToInt(incr["queueing"]),
			candihelper.ToInt(incr["retrying"]), candihelper.ToInt(incr["failure"]), candihelper.ToInt(incr["stopped"]))
		_, err := s.db.Exec(query)
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

func (s *SQLPersistent) toMultiParamQuery(params []string) string {
	var in []string
	for _, taskName := range params {
		in = append(in, "'"+s.queryReplacer.Replace(taskName)+"'")
	}
	return " (" + strings.Join(in, ",") + ") "
}

func (s *SQLPersistent) GetAllConfiguration(ctx context.Context) (cfg []Configuration, err error) {
	query := "SELECT " + s.formatMultiColumnName("key, name, value, is_active") + " FROM " + configurationModelName +
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
	query := "SELECT " + s.formatMultiColumnName("key, name, value, is_active") + " FROM " + configurationModelName +
		` WHERE ` + s.formatColumnName("key") + `='` + s.queryReplacer.Replace(key) + `'`
	err = s.db.QueryRow(query).Scan(&cfg.Key, &cfg.Name, &cfg.Value, &cfg.IsActive)
	return
}

func (s *SQLPersistent) SetConfiguration(cfg *Configuration) (err error) {
	res, err := s.db.Exec(`UPDATE ` + configurationModelName + ` SET 
		` + s.formatColumnName("name") + `='` + s.queryReplacer.Replace(cfg.Name) + `', 
		` + s.formatColumnName("value") + `='` + s.queryReplacer.Replace(cfg.Value) + `', 
		` + s.formatColumnName("is_active") + `=` + candihelper.ToString(cfg.IsActive) + `
		WHERE ` + s.formatColumnName("key") + ` = '` + s.queryReplacer.Replace(cfg.Key) + `'`)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		query := `INSERT INTO ` + configurationModelName + ` (` +
			s.formatColumnName("key, name, value, is_active") + `
			) VALUES (
			'` + s.queryReplacer.Replace(cfg.Key) + `',
			'` + s.queryReplacer.Replace(cfg.Name) + `',
			'` + s.queryReplacer.Replace(cfg.Value) + `',
			` + candihelper.ToString(cfg.IsActive) + `
		)`
		_, err := s.db.Exec(query)
		if err != nil {
			logger.LogE(err.Error())
			return err
		}
	}
	return nil
}
