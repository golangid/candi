package taskqueueworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candiutils"
)

const (
	defaultInterval = "1s"
)

type (
	// Job model
	Job struct {
		ID          string    `bson:"_id" json:"_id"`
		TaskName    string    `bson:"task_name" json:"task_name"`
		Arguments   string    `bson:"arguments" json:"arguments"`
		Retries     int       `bson:"retries" json:"retries"`
		MaxRetry    int       `bson:"max_retry" json:"max_retry"`
		Interval    string    `bson:"interval" json:"interval"`
		CreatedAt   time.Time `bson:"created_at" json:"created_at"`
		FinishedAt  time.Time `bson:"finished_at" json:"finished_at"`
		Status      string    `bson:"status" json:"status"`
		Error       string    `bson:"error" json:"error"`
		TraceID     string    `bson:"traceId" json:"traceId"`
		NextRetryAt string    `bson:"-" json:"-"`
	}

	errorHistory struct {
		Error   string `json:"error"`
		TraceID string `json:"traceID"`
	}
)

// AddJob public function for add new job in same runtime
func AddJob(taskName string, maxRetry int, args []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	task, ok := registeredTask[taskName]
	if !ok {
		var tasks []string
		for taskName := range registeredTask {
			tasks = append(tasks, taskName)
		}
		return fmt.Errorf("task '%s' unregistered, task must one of [%s]", taskName, strings.Join(tasks, ", "))
	}

	if maxRetry <= 0 {
		return errors.New("Max retry must greater than 0")
	}

	var newJob Job
	newJob.ID = uuid.New().String()
	newJob.TaskName = taskName
	newJob.Arguments = string(args)
	newJob.MaxRetry = maxRetry
	newJob.Interval = defaultInterval
	newJob.Status = string(statusQueueing)
	newJob.CreatedAt = time.Now()

	go func(job *Job, workerIndex int) {
		ctx := context.Background()
		queue.PushJob(job)
		persistent.SaveJob(ctx, job)
		broadcastAllToSubscribers(ctx)
		registerJobToWorker(job, workerIndex)
		refreshWorkerNotif <- struct{}{}
	}(&newJob, task.workerIndex)

	return nil
}

// AddJobViaHTTPRequest public function for add new job via http request
func AddJobViaHTTPRequest(ctx context.Context, workerHost string, taskName string, maxRetry int, args []byte) error {
	httpReq := candiutils.NewHTTPRequest(
		candiutils.HTTPRequestSetBreakerName("task_queue_worker_add_job"),
	)

	header := map[string]string{
		candihelper.HeaderContentType: candihelper.HeaderMIMEApplicationJSON,
	}
	reqBody := map[string]interface{}{
		"operationName": "AddJob",
		"variables": map[string]interface{}{
			"taskName": taskName,
			"maxRetry": maxRetry,
			"args":     string(args),
		},
		"query": "mutation AddJob($taskName: String!, $maxRetry: Int!, $args: String!) {\n  add_job(task_name: $taskName, max_retry: $maxRetry, args: $args)\n}\n",
	}
	respBody, _, err := httpReq.Do(ctx, http.MethodPost, workerHost+"/graphql", candihelper.ToBytes(reqBody), header)
	if err != nil {
		return err
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
		return errors.New(respPayload.Errors[0].Message)
	}
	return nil
}

func registerJobToWorker(job *Job, workerIndex int) {
	interval, _ := time.ParseDuration(job.Interval)
	taskIndex := workerIndexTask[workerIndex]
	if taskIndex.activeInterval == nil {
		taskIndex.activeInterval = time.NewTicker(interval)
	} else {
		taskIndex.activeInterval.Reset(interval)
	}
	workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
}
