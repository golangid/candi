package taskqueueworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

type (
	// AddJobRequest request model
	AddJobRequest struct {
		TaskName       string        `json:"task_name"`
		MaxRetry       int           `json:"max_retry"`
		Args           []byte        `json:"args"`
		RetryInterval  time.Duration `json:"retry_interval"`
		CronExpression string        `json:"cron_expression"`

		direct   bool              `json:"-"`
		schedule cronexpr.Schedule `json:"-"`
	}
)

// Validate method
func (a *AddJobRequest) Validate() error {
	if a.CronExpression != "" {
		schedule, err := cronexpr.Parse(a.CronExpression)
		if err != nil {
			return err
		}
		a.schedule = schedule
		return nil
	}

	switch {
	case a.TaskName == "":
		return errors.New("Task name cannot empty")
	case a.MaxRetry <= 0:
		return errors.New("Max retry cannot less or equal than zero")
	case a.RetryInterval < 0:
		return errors.New("Retry interval cannot less than zero")
	}

	return nil
}

// AddJob public function for add new job in same runtime
func AddJob(ctx context.Context, req *AddJobRequest) (jobID string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TaskQueueWorker:AddJob")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("task_name", req.TaskName)

	if err = req.Validate(); err != nil {
		return jobID, err
	}

	if !req.direct && externalWorkerHost != "" {
		return AddJobViaHTTPRequest(ctx, externalWorkerHost, req)
	}

	trace.Log("message", req.Args)

	if engine == nil {
		return jobID, errWorkerInactive
	}
	workerIndex, ok := engine.registeredTaskWorkerIndex[req.TaskName]
	if !ok {
		return jobID, fmt.Errorf("task '%s' unregistered, task must one of [%s]",
			req.TaskName, strings.Join(engine.tasks, ", "))
	}

	var newJob Job
	newJob.TaskName = req.TaskName
	newJob.Arguments = string(req.Args)
	newJob.MaxRetry = req.MaxRetry
	newJob.Interval = defaultInterval.String()
	if req.RetryInterval > 0 {
		newJob.Interval = req.RetryInterval.String()
	}
	if req.CronExpression != "" {
		if totalJob := engine.opt.persistent.CountAllJob(ctx, &Filter{
			TaskName: req.TaskName, MaxRetry: candihelper.WrapPtr(0),
			Statuses: []string{string(StatusQueueing), string(StatusRetrying)},
		}); totalJob > 0 {
			return jobID, fmt.Errorf("there is running cron job in task '%s'", req.TaskName)
		}
		newJob.Interval = req.CronExpression
		newJob.NextRetryAt = req.schedule.Next(time.Now()).Format(time.RFC3339)
	}
	newJob.Status = string(StatusQueueing)
	newJob.CreatedAt = time.Now()
	newJob.direct = req.direct

	ctx = context.WithoutCancel(ctx)
	summary := engine.opt.persistent.Summary().FindDetailSummary(ctx, req.TaskName)
	if summary.IsHold {
		newJob.Status = string(StatusHold)
		newJob.RetryHistories = []RetryHistory{
			{Status: newJob.Status, StartAt: time.Now(), EndAt: time.Now()},
		}
	}

	if err := engine.opt.persistent.SaveJob(ctx, &newJob); err != nil {
		trace.SetError(err)
		logger.LogE(fmt.Sprintf("Cannot save job, error: %s", err.Error()))
		newJob.ID = ""
		err = engine.opt.secondaryPersistent.SaveJob(ctx, &newJob)
		return newJob.ID, err
	}
	trace.SetTag("job_id", newJob.ID)

	engine.opt.persistent.Summary().IncrementSummary(ctx, newJob.TaskName, map[string]int64{
		strings.ToLower(newJob.Status): 1,
	})
	engine.subscriber.broadcastAllToSubscribers(ctx)
	if summary.IsHold || summary.IsLoading {
		return newJob.ID, nil
	}
	if n := engine.opt.queue.PushJob(ctx, &newJob); n <= 1 && len(engine.semaphore[workerIndex-1]) == 0 {
		engine.registerJobToWorker(&newJob)
	}
	if engine.opt.locker.HasBeenLocked(engine.getLockKey(newJob.TaskName)) {
		engine.unlockTask(newJob.TaskName)
	}

	return newJob.ID, nil
}

// AddJobViaHTTPRequest public function for add new job via http request
func AddJobViaHTTPRequest(ctx context.Context, workerHost string, req *AddJobRequest) (jobID string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TaskQueueWorker:AddJobViaHTTPRequest")
	defer trace.Finish()

	if err = req.Validate(); err != nil {
		return jobID, err
	}

	httpReq := candiutils.NewHTTPRequest(
		candiutils.HTTPRequestSetBreakerName("task_queue_worker_add_job"),
		candiutils.HTTPRequestSetClient(&http.Client{
			Timeout: 30 * time.Second,
		}),
	)

	header := map[string]string{
		candihelper.HeaderContentType: candihelper.HeaderMIMEApplicationJSON,
	}

	param := map[string]any{
		"task_name": req.TaskName,
		"max_retry": req.MaxRetry,
		"args":      string(req.Args),
	}
	if req.RetryInterval > 0 {
		param["retry_interval"] = req.RetryInterval.String()
	}

	reqBody := map[string]any{
		"operationName": "addJob",
		"variables": map[string]any{
			"param": param,
		},
		"query": `mutation addJob($param: AddJobInputResolver!) { add_job(param: $param) }`,
	}
	httpResp, err := httpReq.DoRequest(ctx, http.MethodPost, strings.Trim(workerHost, "/")+"/graphql", candihelper.ToBytes(reqBody), header)
	if err != nil {
		return jobID, err
	}

	var respPayload struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Data struct {
			AddJob string `json:"add_job"`
		} `json:"data"`
	}
	json.Unmarshal(httpResp.Bytes(), &respPayload)
	if len(respPayload.Errors) > 0 {
		return jobID, errors.New(respPayload.Errors[0].Message)
	}
	trace.SetTag("job_id", respPayload.Data.AddJob)
	return respPayload.Data.AddJob, nil
}

// AddJobWorkerHandler worker handler, bridging request from another worker for add job, default max retry is 5
func AddJobWorkerHandler(taskName string) types.WorkerHandlerFunc {
	return func(eventContext *candishared.EventContext) error {
		_, err := AddJob(eventContext.Context(), &AddJobRequest{
			TaskName: taskName, Args: eventContext.Message(), MaxRetry: 5,
		})
		return err
	}
}

// GetDetailJob api for get detail job by id
func GetDetailJob(ctx context.Context, jobID string) (Job, error) {
	if engine == nil {
		return Job{}, errWorkerInactive
	}
	return engine.opt.persistent.FindJobByID(ctx, jobID, nil)
}

// RetryJob api for retry job by id
func RetryJob(ctx context.Context, jobID string) error {
	if engine == nil {
		return errWorkerInactive
	}

	job, err := engine.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}

	engine.opt.locker.Unlock(engine.getLockKey(job.TaskName))
	engine.opt.locker.Unlock(engine.getLockKey(job.ID))

	if _, err := cronexpr.Parse(job.Interval); err != nil {
		job.Interval = defaultInterval.String()
	}

	statusBefore := job.Status
	job.Status = string(StatusQueueing)
	if (job.Status == string(StatusFailure)) || (job.Retries >= job.MaxRetry) {
		job.Retries = 0
	}
	matched, affected, err := engine.opt.persistent.UpdateJob(ctx, &Filter{JobID: &job.ID}, map[string]any{
		"status": job.Status, "interval": job.Interval, "retries": job.Retries,
	})
	if err != nil {
		logger.LogE(err.Error())
		return err
	}
	engine.opt.persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]int64{
		statusBefore: -matched,
		job.Status:   affected,
	})

	_, ok := engine.registeredTaskWorkerIndex[job.TaskName]
	if !ok {
		err := errors.New("Task not found")
		logger.LogE(err.Error())
		return err
	}
	engine.opt.queue.PushJob(ctx, &job)
	engine.subscriber.broadcastAllToSubscribers(ctx)
	engine.registerJobToWorker(&job)

	return nil
}

// StopJob api for stop job by id
func StopJob(ctx context.Context, jobID string) error {
	if engine == nil {
		return errWorkerInactive
	}

	job, err := engine.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		return err
	}

	statusBefore := job.Status
	if job.Status == string(StatusRetrying) {
		engine.stopAllJobInTask(job.TaskName)
	}

	job.Status = string(StatusStopped)
	matchedCount, countAffected, err := engine.opt.persistent.UpdateJob(
		ctx, &Filter{JobID: &job.ID, Status: &statusBefore},
		map[string]any{"status": job.Status},
	)
	if err != nil {
		logger.LogE(err.Error())
		return err
	}
	engine.opt.persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]int64{
		job.Status:   countAffected,
		statusBefore: -matchedCount,
	})
	engine.subscriber.broadcastAllToSubscribers(ctx)
	engine.registerNextJob(false, job.TaskName)

	return nil
}

// StreamAllJob api func for stream fetch all job, return total job
func StreamAllJob(ctx context.Context, filter *Filter, streamFunc func(idx, total int, job *Job)) (count int) {
	if engine == nil {
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	perst := engine.opt.persistent
	if filter.secondaryPersistent {
		perst = engine.opt.secondaryPersistent
	}

	count = perst.CountAllJob(ctx, filter)
	if count == 0 || streamFunc == nil {
		return count
	}

	totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
	for filter.Page <= totalPages {
		for i, job := range perst.FindAllJob(ctx, filter) {
			offset := (filter.Page - 1) * filter.Limit
			streamFunc(offset+i, count, &job)
		}
		filter.Page++
	}
	return count
}

// RecalculateSummary func
func RecalculateSummary(ctx context.Context) {
	if engine == nil {
		return
	}

	mapper := make(map[string]TaskSummary, len(engine.tasks))
	for _, taskSummary := range engine.opt.persistent.AggregateAllTaskJob(ctx, &Filter{}) {
		mapper[taskSummary.ID] = taskSummary
	}

	for _, task := range engine.tasks {
		taskSummary, ok := mapper[task]
		if !ok {
			taskSummary.ID = task
		}
		engine.opt.persistent.Summary().UpdateSummary(ctx, taskSummary.ID, map[string]any{
			"success":  taskSummary.Success,
			"queueing": taskSummary.Queueing,
			"retrying": taskSummary.Retrying,
			"failure":  taskSummary.Failure,
			"stopped":  taskSummary.Stopped,
		})
	}
}

// UpdateProgressJob api for update progress job
func UpdateProgressJob(ctx context.Context, jobID string, numProcessed, maxProcess int) error {
	if engine == nil {
		return errWorkerInactive
	}

	if numProcessed > maxProcess {
		return errors.New("Num processed cannot greater than max process")
	}

	job, err := engine.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		return err
	}

	_, _, err = engine.opt.persistent.UpdateJob(ctx, &Filter{
		JobID: &job.ID,
	}, map[string]any{
		"current_progress": numProcessed, "max_progress": maxProcess,
	})
	if err != nil {
		return err
	}

	if len(engine.subscriber.clientJobDetailSubscribers) > 0 {
		engine.globalSemaphore <- struct{}{}
		go func() {
			defer func() { <-engine.globalSemaphore }()
			engine.subscriber.broadcastJobDetail(context.WithoutCancel(ctx))
		}()
	}
	return nil
}
