package taskqueueworker

import (
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	cronexpr "github.com/golangid/candi/candiutils/cronparser"
)

const (
	jobModelName           = "task_queue_worker_jobs"
	jobSummaryModelName    = "task_queue_worker_job_summaries"
	jobHistoryModel        = "task_queue_worker_job_histories"
	configurationModelName = "task_queue_worker_configurations"
)

// Filter type
type Filter struct {
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
	MaxRetry            *int       `json:"maxRetry,omitempty"`
	secondaryPersistent bool       `json:"-"`
}

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

// TaskSummary model
type TaskSummary struct {
	ID             string     `bson:"_id"`
	TaskName       string     `bson:"task_name"`
	Success        int        `bson:"success"`
	Queueing       int        `bson:"queueing"`
	Retrying       int        `bson:"retrying"`
	Failure        int        `bson:"failure"`
	Stopped        int        `bson:"stopped"`
	Hold           int        `bson:"hold"`
	IsLoading      bool       `bson:"is_loading"`
	IsHold         bool       `bson:"is_hold"`
	LoadingMessage string     `bson:"loading_message"`
	Config         TaskConfig `bson:"config"`
}

type TaskConfig struct {
	HoldUnholdSwitchInterval string `bson:"hold_unhold_switch_interval" json:"hold_unhold_switch_interval"`
	NextHoldUnhold           string `bson:"next_hold_unhold" json:"next_hold_unhold"`
}

// CountTotalJob method
func (s *TaskSummary) CountTotalJob() int {
	return normalizeCount(s.Success) + normalizeCount(s.Queueing) + normalizeCount(s.Retrying) +
		normalizeCount(s.Failure) + normalizeCount(s.Stopped) + normalizeCount(s.Hold)
}

// ToSummaryDetail method
func (s *TaskSummary) ToSummaryDetail() (detail SummaryDetail) {
	detail.Failure = normalizeCount(s.Failure)
	detail.Retrying = normalizeCount(s.Retrying)
	detail.Success = normalizeCount(s.Success)
	detail.Queueing = normalizeCount(s.Queueing)
	detail.Stopped = normalizeCount(s.Stopped)
	detail.Hold = normalizeCount(s.Hold)
	return
}

// ToTaskResolver method
func (s *TaskSummary) ToTaskResolver() (res TaskResolver) {
	regTask, ok := engine.registeredTaskWorkerIndex[s.TaskName]
	if !ok {
		return
	}

	res = TaskResolver{
		Name:           s.TaskName,
		ModuleName:     engine.runningWorkerIndexTask[regTask].moduleName,
		TotalJobs:      s.CountTotalJob(),
		IsLoading:      s.IsLoading,
		IsHold:         s.IsHold,
		LoadingMessage: s.LoadingMessage,
	}
	res.Detail = s.ToSummaryDetail()
	return
}

// ToMapResult method
func (s *TaskSummary) ToMapResult() map[string]int {
	return map[string]int{
		strings.ToUpper(string(StatusFailure)):  s.Failure,
		strings.ToUpper(string(StatusRetrying)): s.Retrying,
		strings.ToUpper(string(StatusSuccess)):  s.Success,
		strings.ToUpper(string(StatusQueueing)): s.Queueing,
		strings.ToUpper(string(StatusStopped)):  s.Stopped,
		strings.ToUpper(string(StatusHold)):     s.Hold,
	}
}

// SetValue method
func (s *TaskSummary) SetValue(source map[string]int) {
	s.Failure = source[strings.ToUpper(string(StatusFailure))]
	s.Retrying = source[strings.ToUpper(string(StatusRetrying))]
	s.Success = source[strings.ToUpper(string(StatusSuccess))]
	s.Queueing = source[strings.ToUpper(string(StatusQueueing))]
	s.Stopped = source[strings.ToUpper(string(StatusStopped))]
	s.Hold = source[strings.ToUpper(string(StatusHold))]
}

// ApplyFilterStatus apply with filter status
func (s *TaskSummary) ApplyFilterStatus(statuses []string) {
	if len(statuses) == 0 {
		return
	}
	mapRes := s.ToMapResult()
	newCount := map[string]int{}
	for _, status := range statuses {
		newCount[strings.ToUpper(status)] = mapRes[strings.ToUpper(status)]
	}
	s.SetValue(newCount)
}

func (s TaskSummary) GetColumnName() []string {
	return []string{"id", "success", "queueing", "retrying", "failure", "stopped", "is_loading", "is_hold", "hold", "loading_message"}
}

func (s *TaskSummary) Scan(scanner interface{ Scan(...any) error }) error {
	return scanner.Scan(&s.TaskName, &s.Success, &s.Queueing, &s.Retrying,
		&s.Failure, &s.Stopped, &s.IsLoading, &s.IsHold, &s.Hold, &s.LoadingMessage)
}

func (s *TaskSummary) ToArgs(val map[string]any) (args []any) {
	return []any{
		s.TaskName, candihelper.ToInt(val["success"]), candihelper.ToInt(val["queueing"]), candihelper.ToInt(val["retrying"]),
		candihelper.ToInt(val["failure"]), candihelper.ToInt(val["stopped"]), val["is_loading"],
		val["is_hold"], val["hold"], val["loading_message"],
	}
}

// Job model
type Job struct {
	ID              string         `bson:"_id" json:"_id"`
	TaskName        string         `bson:"task_name" json:"task_name"`
	Arguments       string         `bson:"arguments" json:"arguments"`
	Retries         int            `bson:"retries" json:"retries"`
	MaxRetry        int            `bson:"max_retry" json:"max_retry"`
	Interval        string         `bson:"interval" json:"interval"`
	CreatedAt       time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `bson:"updated_at" json:"updated_at"`
	FinishedAt      time.Time      `bson:"finished_at" json:"finished_at"`
	NextRunningAt   time.Time      `bson:"next_running_at" json:"next_running_at"`
	Status          string         `bson:"status" json:"status"`
	Error           string         `bson:"error" json:"error"`
	ErrorStack      string         `bson:"-" json:"error_stack"`
	Result          string         `bson:"result" json:"result"`
	TraceID         string         `bson:"trace_id" json:"trace_id"`
	CurrentProgress int64          `bson:"current_progress" json:"current_progress"`
	MaxProgress     int64          `bson:"max_progress" json:"max_progress"`
	RetryHistories  []RetryHistory `bson:"retry_histories" json:"retry_histories"`

	direct   bool              `bson:"-" json:"-"`
	schedule cronexpr.Schedule `bson:"-" json:"-"`
}

// RetryHistory model
type RetryHistory struct {
	ErrorStack string    `bson:"error_stack" json:"error_stack"`
	Status     string    `bson:"status" json:"status"`
	Error      string    `bson:"error" json:"error"`
	Result     string    `bson:"result" json:"result"`
	TraceID    string    `bson:"trace_id" json:"trace_id"`
	StartAt    time.Time `bson:"start_at" json:"start_at"`
	EndAt      time.Time `bson:"end_at" json:"end_at"`
}

func (job *Job) toMap() map[string]any {
	return map[string]any{
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

func (j *Job) IsCronMode() bool {
	_, err := time.ParseDuration(j.Interval)
	return err != nil
}

func (j *Job) ParseNextRunningInterval() (interval time.Duration, err error) {
	if !j.NextRunningAt.IsZero() && j.NextRunningAt.After(time.Now()) {
		interval = j.NextRunningAt.Sub(time.Now())
		return
	}
	interval, err = time.ParseDuration(j.Interval)
	if err != nil || interval <= 0 {
		schedule, err := cronexpr.Parse(j.Interval)
		if err != nil {
			return interval, err
		}
		interval = schedule.NextInterval(time.Now())
	}
	j.NextRunningAt = time.Now().Add(interval)
	return interval, nil
}

// Configuration model
type Configuration struct {
	Key      string `bson:"key" json:"key"`
	Name     string `bson:"name" json:"name"`
	Value    string `bson:"value" json:"value"`
	IsActive bool   `bson:"is_active" json:"is_active"`
}

type Configurations []Configuration

// ToMap method
func (c Configurations) ToMap() map[string]string {
	mp := make(map[string]string, len(c))
	for _, cfg := range c {
		mp[cfg.Key] = cfg.Value
	}
	return mp
}
