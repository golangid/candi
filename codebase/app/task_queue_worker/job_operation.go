package taskqueueworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		return errors.New("Retry interval cannot less or equal than zero")
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
		trace.SetError(err)
		trace.Finish()
	}()

	trace.SetTag("task_name", job.TaskName)

	if err = job.Validate(); err != nil {
		return jobID, err
	}

	task, ok := registeredTask[job.TaskName]
	if !ok {
		var tasks []string
		for taskName := range registeredTask {
			tasks = append(tasks, taskName)
		}
		return jobID, fmt.Errorf("task '%s' unregistered, task must one of [%s]", job.TaskName, strings.Join(tasks, ", "))
	}

	if job.MaxRetry <= 0 {
		return jobID, errors.New("Max retry must greater than 0")
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

	go func(job *Job, workerIndex int) {
		ctx := context.Background()
		queue.PushJob(ctx, job)
		persistent.SaveJob(ctx, job)
		broadcastAllToSubscribers(ctx)
		registerJobToWorker(job, workerIndex)
		refreshWorkerNotif <- struct{}{}
	}(&newJob, task.workerIndex)

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
	job.Interval = defaultInterval.String()
	job.Status = string(statusQueueing)
	if (job.Status == string(statusFailure)) || (job.Retries >= job.MaxRetry) {
		job.Retries = 0
	}
	persistent.UpdateJob(ctx, Filter{JobID: &job.ID}, map[string]interface{}{
		"status": statusQueueing, "interval": job.Interval, "retries": job.Retries,
	})

	task := registeredTask[job.TaskName]
	queue.PushJob(ctx, job)
	broadcastAllToSubscribers(ctx)
	registerJobToWorker(job, task.workerIndex)

	go func(job *Job) { refreshWorkerNotif <- struct{}{} }(job)
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

	if job.Status == string(statusRetrying) {
		stopAllJobInTask(job.TaskName)
	}

	persistent.UpdateJob(ctx, Filter{JobID: &job.ID}, map[string]interface{}{"status": statusStopped})
	broadcastAllToSubscribers(ctx)

	return nil
}
