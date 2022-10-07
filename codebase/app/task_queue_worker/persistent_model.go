package taskqueueworker

import (
	"strings"
	"time"
)

const (
	jobModelName           = "task_queue_worker_jobs"
	jobSummaryModelName    = "task_queue_worker_job_summaries"
	configurationModelName = "task_queue_worker_configurations"
)

type (
	// Filter type
	Filter struct {
		Page                int        `json:"page"`
		Limit               int        `json:"limit"`
		Sort                string     `json:"sort,omitempty"`
		TaskName            string     `json:"taskName,omitempty"`
		TaskNameList        []string   `json:"taskNameList,omitempty"`
		ExcludeTaskNameList []string   `json:"excludeTaskNameList,omitempty"`
		Search              *string    `json:"search,omitempty"`
		JobID               *string    `json:"jobID,omitempty"`
		Status              *string    `json:"status,omitempty"`
		Statuses            []string   `json:"statuses,omitempty"`
		ExcludeStatus       []string   `json:"excludeStatus,omitempty"`
		ShowAll             bool       `json:"showAll,omitempty"`
		ShowHistories       *bool      `json:"showHistories,omitempty"`
		StartDate           string     `json:"startDate,omitempty"`
		EndDate             string     `json:"endDate,omitempty"`
		BeforeCreatedAt     *time.Time `json:"beforeCreatedAt,omitempty"`
		Count               int        `json:"count,omitempty"`
	}

	// TaskSummary model
	TaskSummary struct {
		ID        string `bson:"_id"`
		TaskName  string `bson:"task_name"`
		Success   int    `bson:"success"`
		Queueing  int    `bson:"queueing"`
		Retrying  int    `bson:"retrying"`
		Failure   int    `bson:"failure"`
		Stopped   int    `bson:"stopped"`
		IsLoading bool   `bson:"is_loading"`
	}

	// Job model
	Job struct {
		ID             string         `bson:"_id" json:"_id"`
		TaskName       string         `bson:"task_name" json:"task_name"`
		Arguments      string         `bson:"arguments" json:"arguments"`
		Retries        int            `bson:"retries" json:"retries"`
		MaxRetry       int            `bson:"max_retry" json:"max_retry"`
		Interval       string         `bson:"interval" json:"interval"`
		CreatedAt      time.Time      `bson:"created_at" json:"created_at"`
		UpdatedAt      time.Time      `bson:"updated_at" json:"updated_at"`
		FinishedAt     time.Time      `bson:"finished_at" json:"finished_at"`
		Status         string         `bson:"status" json:"status"`
		Error          string         `bson:"error" json:"error"`
		ErrorStack     string         `bson:"-" json:"error_stack"`
		TraceID        string         `bson:"trace_id" json:"trace_id"`
		RetryHistories []RetryHistory `bson:"retry_histories" json:"retry_histories"`
		NextRetryAt    string         `bson:"-" json:"-"`
		direct         bool           `bson:"-" json:"-"`
	}

	// RetryHistory model
	RetryHistory struct {
		ErrorStack string    `bson:"error_stack" json:"error_stack"`
		Status     string    `bson:"status" json:"status"`
		Error      string    `bson:"error" json:"error"`
		TraceID    string    `bson:"trace_id" json:"trace_id"`
		StartAt    time.Time `bson:"start_at" json:"start_at"`
		EndAt      time.Time `bson:"end_at" json:"end_at"`
	}

	// Configuration model
	Configuration struct {
		Key      string `bson:"key" json:"key"`
		Name     string `bson:"name" json:"name"`
		Value    string `bson:"value" json:"value"`
		IsActive bool   `bson:"is_active" json:"is_active"`
	}

	Configurations []Configuration
)

// CalculateOffset method
func (f *Filter) CalculateOffset() int {
	return (f.Page - 1) * f.Limit
}

// ParseStartEndDate method
func (f *Filter) ParseStartEndDate() (startDate, endDate time.Time) {

	startDate, _ = time.Parse(time.RFC3339, f.StartDate)
	endDate, _ = time.Parse(time.RFC3339, f.EndDate)

	return
}

// CountTotalJob method
func (s *TaskSummary) CountTotalJob() int {
	return s.Success + s.Queueing + s.Retrying + s.Failure + s.Stopped
}

// ToSummaryDetail method
func (s *TaskSummary) ToSummaryDetail() (detail SummaryDetail) {
	detail.Failure = s.Failure
	detail.Retrying = s.Retrying
	detail.Success = s.Success
	detail.Queueing = s.Queueing
	detail.Stopped = s.Stopped
	return
}

// ToMapResult method
func (s *TaskSummary) ToMapResult() map[string]int {
	return map[string]int{
		strings.ToUpper(string(statusFailure)):  s.Failure,
		strings.ToUpper(string(statusRetrying)): s.Retrying,
		strings.ToUpper(string(statusSuccess)):  s.Success,
		strings.ToUpper(string(statusQueueing)): s.Queueing,
		strings.ToUpper(string(statusStopped)):  s.Stopped,
	}
}

// SetValue method
func (s *TaskSummary) SetValue(source map[string]int) {
	s.Failure = source[strings.ToUpper(string(statusFailure))]
	s.Retrying = source[strings.ToUpper(string(statusRetrying))]
	s.Success = source[strings.ToUpper(string(statusSuccess))]
	s.Queueing = source[strings.ToUpper(string(statusQueueing))]
	s.Stopped = source[strings.ToUpper(string(statusStopped))]
}

func (job *Job) toMap() map[string]interface{} {
	return map[string]interface{}{
		"task_name":   job.TaskName,
		"arguments":   job.Arguments,
		"retries":     job.Retries,
		"max_retry":   job.MaxRetry,
		"interval":    job.Interval,
		"updated_at":  job.UpdatedAt,
		"finished_at": job.FinishedAt,
		"status":      job.Status,
		"error":       job.Error,
		"trace_id":    job.TraceID,
	}
}

// ToMap method
func (c Configurations) ToMap() map[string]string {
	mp := make(map[string]string, len(c))
	for _, cfg := range c {
		mp[cfg.Key] = cfg.Value
	}
	return mp
}
