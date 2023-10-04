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
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

type (
	// AddJobRequest request model
	AddJobRequest struct {
		TaskName      string        `json:"task_name"`
		MaxRetry      int           `json:"max_retry"`
		Args          []byte        `json:"args"`
		RetryInterval time.Duration `json:"retry_interval"`
		direct        bool
	}
)

// Validate method
func (a *AddJobRequest) Validate() error {

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
	newJob.Status = string(StatusQueueing)
	newJob.CreatedAt = time.Now()
	newJob.direct = req.direct

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
	engine.subscriber.broadcastAllToSubscribers(context.Background())
	if n := engine.opt.queue.PushJob(ctx, &newJob); n <= 1 && len(engine.semaphore[workerIndex-1]) == 0 {
		engine.registerJobToWorker(&newJob)
	}
	if engine.opt.locker.HasBeenLocked(engine.getLockKey(newJob.TaskName)) {
		engine.checkForUnlockTask(newJob.TaskName)
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

	param := map[string]interface{}{
		"task_name": req.TaskName,
		"max_retry": req.MaxRetry,
		"args":      string(req.Args),
	}
	if req.RetryInterval > 0 {
		param["retry_interval"] = req.RetryInterval.String()
	}

	reqBody := map[string]interface{}{
		"operationName": "addJob",
		"variables": map[string]interface{}{
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

	statusBefore := job.Status
	job.Interval = defaultInterval.String()
	job.Status = string(StatusQueueing)
	if (job.Status == string(StatusFailure)) || (job.Retries >= job.MaxRetry) {
		job.Retries = 0
	}
	matched, affected, err := engine.opt.persistent.UpdateJob(ctx, &Filter{JobID: &job.ID}, map[string]interface{}{
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

	if ctx.Err() != nil {
		ctx = context.Background()
	}

	job, err := engine.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		return err
	}

	statusBefore := job.Status
	if job.Status == string(StatusRetrying) {
		go engine.stopAllJobInTask(job.TaskName)
	}

	job.Status = string(StatusStopped)
	matchedCount, countAffected, err := engine.opt.persistent.UpdateJob(
		ctx, &Filter{JobID: &job.ID, Status: &statusBefore},
		map[string]interface{}{"status": job.Status},
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

// StreamAllJob api func for stream fetch all job
func StreamAllJob(ctx context.Context, filter *Filter, streamFunc func(job *Job)) (count int) {
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
	if count == 0 {
		return
	}

	totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
	for filter.Page <= totalPages {
		for _, job := range perst.FindAllJob(ctx, filter) {
			streamFunc(&job)
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
		engine.opt.persistent.Summary().UpdateSummary(ctx, taskSummary.ID, map[string]interface{}{
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
	}, map[string]interface{}{
		"current_progress": numProcessed, "max_progress": maxProcess,
	})
	if err != nil {
		return err
	}

	if len(engine.subscriber.clientJobDetailSubscribers) > 0 {
		engine.globalSemaphore <- struct{}{}
		go func() {
			defer func() { <-engine.globalSemaphore }()
			engine.subscriber.broadcastJobDetail(context.Background())
		}()
	}
	return nil
}
