package taskqueueworker

import (
	"context"
	"math"
	"time"

	"github.com/golangid/candi/candihelper"
	cronexpr "github.com/golangid/candi/candiutils/cronparser"
)

type (
	// DashboardResolver resolver
	DashboardResolver struct {
		Banner           string
		Tagline          string
		Version          string
		GoVersion        string
		StartAt          string
		BuildNumber      string
		Config           ConfigResolver
		MemoryStatistics MemstatsResolver
		DependencyHealth struct {
			Persistent *string
			Queue      *string
		}
		DependencyDetail struct {
			PersistentType         string
			QueueType              string
			UseSecondaryPersistent bool
		}
	}
	// MemstatsResolver resolver
	MemstatsResolver struct {
		Alloc         uint64
		TotalAlloc    uint64
		NumGC         int
		NumGoroutines int
	}
	// MetaTaskResolver meta resolver
	MetaTaskResolver struct {
		Page                  int
		Limit                 int
		TotalRecords          int
		TotalPages            int
		IsCloseSession        bool
		TotalClientSubscriber int
		ClientID              string
	}
	// TaskResolver resolver
	TaskResolver struct {
		Name           string
		ModuleName     string
		TotalJobs      int
		IsLoading      bool
		LoadingMessage string
		IsHold         bool
		Detail         SummaryDetail
	}
	// TaskListResolver resolver
	TaskListResolver struct {
		Meta MetaTaskResolver
		Data []TaskResolver
	}
	// RestoreSecondaryResolver resolver
	RestoreSecondaryResolver struct {
		TotalData int
		Message   string
	}

	// MetaJobList resolver
	MetaJobList struct {
		Page              int
		Limit             int
		TotalRecords      int
		TotalPages        int
		Message           string
		IsCloseSession    bool
		IsLoading         bool
		IsFreezeBroadcast bool
		IsHold            bool
		Detail            SummaryDetail
	}

	// SummaryDetail type
	SummaryDetail struct {
		Failure, Retrying, Success, Queueing, Stopped, Hold int
	}

	// JobListResolver resolver
	JobListResolver struct {
		Meta MetaJobList
		Data []JobResolver
	}

	// JobResolver resolver
	JobResolver struct {
		ID              string
		TaskName        string
		Arguments       string
		Retries         int
		MaxRetry        int
		Interval        string
		CreatedAt       string
		FinishedAt      string
		Status          string
		Error           string
		Result          string
		ErrorStack      string
		TraceID         string
		RetryHistories  []RetryHistory
		NextRetryAt     string
		CurrentProgress int
		MaxProgress     int
		Meta            struct {
			IsCloseSession   bool
			Page             int
			TotalHistory     int
			IsShowMoreArgs   bool
			IsShowMoreError  bool
			IsShowMoreResult bool
		}
	}

	// ClientSubscriber model
	ClientSubscriber struct {
		ClientID   string
		PageName   string
		PageFilter string
	}

	// ConfigResolver resolver
	ConfigResolver struct {
		WithPersistent bool
	}

	// GetAllJobInputResolver resolver
	GetAllJobInputResolver struct {
		TaskName  *string
		Page      *int
		Limit     *int
		Search    *string
		JobID     *string
		Statuses  *[]string
		StartDate *string
		EndDate   *string
	}

	// GetAllJobHistoryInputResolver resolver
	GetAllJobHistoryInputResolver struct {
		Page      *int
		Limit     *int
		StartDate *string
		EndDate   *string
		JobID     string
	}

	// ConfigurationResolver resolver
	ConfigurationResolver struct {
		Key      string
		Name     string
		Value    string
		IsActive bool
	}

	// FilterMutateJobInputResolver resolver
	FilterMutateJobInputResolver struct {
		TaskName  string
		Search    *string
		JobID     *string
		Statuses  []string
		StartDate *string
		EndDate   *string
	}
)

// ToFilter method
func (i *GetAllJobInputResolver) ToFilter() (filter Filter) {

	filter = Filter{
		Page: 1, Limit: 10,
		Search: i.Search, TaskName: candihelper.PtrToString(i.TaskName),
		JobID: i.JobID,
	}

	if i.Page != nil && *i.Page > 0 {
		filter.Page = *i.Page
	}
	if i.Limit != nil && *i.Limit > 0 {
		filter.Limit = *i.Limit
	}
	if i.Statuses != nil {
		filter.Statuses = *i.Statuses
	}

	if i.StartDate != nil {
		filter.StartDate = *i.StartDate
	}
	if i.EndDate != nil {
		filter.EndDate = *i.EndDate
	}

	return
}

// ToFilter method
func (i *GetAllJobHistoryInputResolver) ToFilter() (filter Filter) {

	filter = Filter{
		Page: 1, Limit: 10,
	}

	if i.Page != nil && *i.Page > 0 {
		filter.Page = *i.Page
	}
	if i.Limit != nil && *i.Limit > 0 {
		filter.Limit = *i.Limit
	}
	if i.StartDate != nil {
		filter.StartDate = *i.StartDate
	}
	if i.EndDate != nil {
		filter.EndDate = *i.EndDate
	}
	return
}

// ToFilter method
func (i *FilterMutateJobInputResolver) ToFilter() (filter Filter) {

	filter = Filter{
		Page: 1, Limit: 10,
		Search: i.Search, TaskName: i.TaskName,
		JobID: i.JobID,
	}

	filter.Page = 1
	filter.Limit = 10
	filter.Statuses = i.Statuses

	if i.StartDate != nil {
		filter.StartDate = *i.StartDate
	}
	if i.EndDate != nil {
		filter.EndDate = *i.EndDate
	}

	return
}

func (m *MetaTaskResolver) CalculatePage() {
	m.TotalPages = int(math.Ceil(float64(m.TotalRecords) / float64(m.Limit)))
}

func (j *JobResolver) ParseFromJob(job *Job, maxArgsLength int) {
	j.ID = job.ID
	j.TaskName = job.TaskName

	j.Arguments = job.Arguments
	j.Error = job.Error
	j.Result = job.Result
	if maxArgsLength > 0 {
		if len(job.Arguments) > maxArgsLength {
			j.Arguments = job.Arguments[:maxArgsLength]
			j.Meta.IsShowMoreArgs = true
		}
		if len(job.Error) > maxArgsLength {
			j.Error = job.Error[:maxArgsLength]
			j.Meta.IsShowMoreError = true
		}
		if len(job.Result) > maxArgsLength {
			j.Result = job.Result[:maxArgsLength]
			j.Meta.IsShowMoreResult = true
		}
	}
	j.Retries = job.Retries
	j.MaxRetry = job.MaxRetry
	j.Interval = job.Interval
	j.Status = job.Status
	j.ErrorStack = job.ErrorStack
	j.TraceID = job.TraceID
	j.RetryHistories = job.RetryHistories
	j.NextRetryAt = job.NextRetryAt
	j.CurrentProgress = job.CurrentProgress
	j.MaxProgress = job.MaxProgress
	j.RetryHistories = job.RetryHistories
	if job.Status == string(StatusSuccess) {
		j.Error = ""
	}

	if job.Status == string(StatusQueueing) {
		if delay, err := time.ParseDuration(job.Interval); err == nil {
			j.NextRetryAt = time.Now().Add(delay).In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
		} else if schedule, err := cronexpr.Parse(job.Interval); err == nil {
			j.NextRetryAt = schedule.Next(time.Now()).Format(time.RFC3339)
		}
	}
	j.CreatedAt = job.CreatedAt.In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
	if !job.FinishedAt.IsZero() {
		j.FinishedAt = job.FinishedAt.In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
	}
	if job.MaxRetry > 0 && job.Retries > job.MaxRetry {
		j.Retries = job.MaxRetry
	}

	for i := range job.RetryHistories {
		job.RetryHistories[i].StartAt = job.RetryHistories[i].StartAt.In(candihelper.AsiaJakartaLocalTime)
		job.RetryHistories[i].EndAt = job.RetryHistories[i].EndAt.In(candihelper.AsiaJakartaLocalTime)
	}
}

func (j *JobListResolver) GetAllJob(ctx context.Context, filter *Filter) {
	var detailSummary TaskSummary
	if candihelper.PtrToString(filter.Search) != "" ||
		candihelper.PtrToString(filter.JobID) != "" ||
		(filter.StartDate != "" && filter.EndDate != "") {
		taskDetailSummary := engine.opt.persistent.AggregateAllTaskJob(ctx, filter)
		if len(taskDetailSummary) > 0 {
			detailSummary = taskDetailSummary[0]
		}
	} else {
		detailSummary = engine.opt.persistent.Summary().FindDetailSummary(ctx, filter.TaskName)
		detailSummary.ApplyFilterStatus(filter.Statuses)
	}

	detail := detailSummary.ToSummaryDetail()
	j.Meta.Detail.Failure = detail.Failure
	j.Meta.Detail.Retrying = detail.Retrying
	j.Meta.Detail.Success = detail.Success
	j.Meta.Detail.Queueing = detail.Queueing
	j.Meta.Detail.Stopped = detail.Stopped
	j.Meta.Detail.Hold = detail.Hold
	j.Meta.TotalRecords = detailSummary.CountTotalJob()
	j.Meta.IsHold = detailSummary.IsHold
	j.Meta.Message = detailSummary.LoadingMessage
	j.Meta.Page, j.Meta.Limit = filter.Page, filter.Limit
	j.Meta.TotalPages = int(math.Ceil(float64(j.Meta.TotalRecords) / float64(j.Meta.Limit)))

	jobs := engine.opt.persistent.FindAllJob(ctx, filter)
	j.Data = make([]JobResolver, len(jobs))
	for i, job := range jobs {
		j.Data[i].ParseFromJob(&job, 100)
	}
}
