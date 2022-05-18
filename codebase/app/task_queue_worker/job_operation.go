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
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
)

type (
	// AddJobRequest request model
	AddJobRequest struct {
		TaskName      string        `json:"task_name"`
		MaxRetry      int           `json:"max_retry"`
		Args          []byte        `json:"args"`
		RetryInterval time.Duration `json:"retry_interval"`
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
func AddJob(ctx context.Context, job *AddJobRequest) (jobID string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "TaskQueueWorker:AddJob")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("task_name", job.TaskName)

	if err = job.Validate(); err != nil {
		return jobID, err
	}

	task, ok := registeredTask[job.TaskName]
	if !ok {
		return jobID, fmt.Errorf("task '%s' unregistered, task must one of [%s]", job.TaskName, strings.Join(tasks, ", "))
	}

	var newJob Job
	newJob.ID = uuid.New().String()
	newJob.TaskName = job.TaskName
	newJob.Arguments = string(job.Args)
	newJob.MaxRetry = job.MaxRetry
	newJob.Interval = defaultInterval.String()
	if job.RetryInterval > 0 {
		newJob.Interval = job.RetryInterval.String()
	}
	newJob.Status = string(statusQueueing)
	newJob.CreatedAt = time.Now()

	trace.Log("new_job_id", newJob.ID)

	semaphoreAddJob <- struct{}{}
	go func(ctx context.Context, job *Job, workerIndex int) {
		defer func() {
			refreshWorkerNotif <- struct{}{}
			<-semaphoreAddJob
		}()

		persistent.SaveJob(ctx, job)
		queue.PushJob(ctx, job)
		persistent.IncrementSummary(ctx, job.TaskName, map[string]interface{}{
			strings.ToLower(job.Status): 1,
		})
		broadcastAllToSubscribers(ctx)
		registerJobToWorker(job, workerIndex)
	}(context.Background(), &newJob, task.workerIndex)

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
		"query": `mutation addJob($param: AddJobInputResolver!) {
			add_job(param: $param)
		}`,
	}
	respBody, _, err := httpReq.Do(ctx, http.MethodPost, workerHost+"/graphql", candihelper.ToBytes(reqBody), header)
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
	json.Unmarshal(respBody, &respPayload)
	if len(respPayload.Errors) > 0 {
		return jobID, errors.New(respPayload.Errors[0].Message)
	}
	return respPayload.Data.AddJob, nil
}

// GetDetailJob api for get detail job by id
func GetDetailJob(ctx context.Context, jobID string) (*Job, error) {
	return persistent.FindJobByID(ctx, jobID)
}

// RetryJob api for retry job by id
func RetryJob(ctx context.Context, jobID string) error {
	job, err := persistent.FindJobByID(ctx, jobID)
	if err != nil {
		return err
	}

	statusBefore := job.Status
	job.Interval = defaultInterval.String()
	job.Status = string(statusQueueing)
	if (job.Status == string(statusFailure)) || (job.Retries >= job.MaxRetry) {
		job.Retries = 0
	}
	matched, affected, _ := persistent.UpdateJob(ctx, &Filter{JobID: &job.ID}, map[string]interface{}{
		"status": job.Status, "interval": job.Interval, "retries": job.Retries,
	})
	persistent.IncrementSummary(ctx, job.TaskName, map[string]interface{}{
		statusBefore: -1 * matched,
		job.Status:   affected,
	})

	task := registeredTask[job.TaskName]
	queue.PushJob(ctx, job)
	broadcastAllToSubscribers(ctx)
	registerJobToWorker(job, task.workerIndex)

	go func() { refreshWorkerNotif <- struct{}{} }()
	return nil
}

// StopJob api for stop job by id
func StopJob(ctx context.Context, jobID string) error {

	if ctx.Err() != nil {
		ctx = context.Background()
	}

	job, err := persistent.FindJobByID(ctx, jobID)
	if err != nil {
		return err
	}

	statusBefore := job.Status
	if job.Status == string(statusRetrying) {
		stopAllJobInTask(job.TaskName)
	}

	job.Status = string(statusStopped)
	matchedCount, countAffected, err := persistent.UpdateJob(
		ctx, &Filter{JobID: &job.ID, Status: &statusBefore},
		map[string]interface{}{"status": job.Status},
	)
	if err != nil {
		return err
	}
	persistent.IncrementSummary(ctx, job.TaskName, map[string]interface{}{
		job.Status:   countAffected,
		statusBefore: -1 * matchedCount,
	})
	broadcastAllToSubscribers(ctx)

	return nil
}

// StreamAllJob api func for stream fetch all job
func StreamAllJob(ctx context.Context, filter *Filter, streamFunc func(job *Job)) {

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	count := persistent.CountAllJob(ctx, filter)
	if count == 0 {
		return
	}

	totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
	for filter.Page <= totalPages {
		for _, job := range persistent.FindAllJob(ctx, filter) {
			streamFunc(&job)
		}
		filter.Page++
	}
}

// RecalculateSummary func
func RecalculateSummary(ctx context.Context) {

	mapper := make(map[string]TaskSummary, len(tasks))
	for _, taskSummary := range persistent.AggregateAllTaskJob(ctx, &Filter{}) {
		mapper[taskSummary.ID] = taskSummary
	}

	for _, task := range tasks {
		taskSummary, ok := mapper[task]
		if !ok {
			taskSummary.ID = task
		}
		persistent.UpdateSummary(ctx, taskSummary.ID, map[string]interface{}{
			"success":  taskSummary.Success,
			"queueing": taskSummary.Queueing,
			"retrying": taskSummary.Retrying,
			"failure":  taskSummary.Failure,
			"stopped":  taskSummary.Stopped,
		})
	}
}
