package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
	"github.com/google/uuid"
)

// Job model
type Job struct {
	ID             string         `json:"id"`
	TaskName       string         `json:"task_name"`
	Args           []byte         `json:"args"`
	Retries        int            `json:"retries"`
	MaxRetry       int            `json:"max_retry"`
	Interval       string         `json:"interval"`
	ErrorHistories []errorHistory `json:"error_histories"`
}

type errorHistory struct {
	Error   string `json:"error"`
	TraceID string `json:"traceID"`
}

// AddJob public function
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

	var newJob Job
	newJob.ID = uuid.New().String()
	newJob.TaskName = taskName
	newJob.Args = args
	newJob.MaxRetry = maxRetry
	newJob.Interval = "1s"

	isRefresh := workerIndexTask[task.workerIndex].activeInterval == nil
	registerJobToWorker(&newJob, task.workerIndex)

	queue.PushJob(&newJob)

	if isRefresh {
		refreshWorkerNotif <- struct{}{}
	}
	return nil
}

func registerJobToWorker(job *Job, workerIndex int) {
	interval, _ := time.ParseDuration(job.Interval)
	taskIndex := workerIndexTask[workerIndex]
	taskIndex.activeInterval = time.NewTicker(interval)
	workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
}

func execJob(workerIndex int) {
	trace := tracer.StartTrace(context.Background(), "TaskQueueWorker")
	defer trace.Finish()
	ctx := trace.Context()

	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}
		refreshWorkerNotif <- struct{}{}
		logger.LogGreen(tracer.GetTraceURL(ctx))
	}()

	taskIndex := workerIndexTask[workerIndex]
	taskIndex.activeInterval.Stop()
	taskIndex.activeInterval = nil

	job := queue.PopJob(taskIndex.taskName)
	job.Retries++

	tags := trace.Tags()
	tags["job_id"] = job.ID
	tags["task_name"] = job.TaskName
	tags["job_args"] = string(job.Args)
	tags["retries"] = job.Retries
	tags["max_retry"] = job.MaxRetry

	nextJob := queue.NextJob(taskIndex.taskName)
	if nextJob != nil {
		registerJobToWorker(nextJob, workerIndex)
	}

	log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s'\x1b[0m", job.TaskName)
	if err := registeredTask[job.TaskName].handlerFunc(ctx, job.Args); err != nil {
		job.ErrorHistories = append(job.ErrorHistories, errorHistory{
			Error:   err.Error(),
			TraceID: tracer.GetTraceID(ctx),
		})
		tags["job_error_histories"] = job.ErrorHistories
		switch e := err.(type) {
		case *ErrorRetrier:
			if job.Retries >= job.MaxRetry {
				fmt.Printf("\x1b[31;1mTaskQueueWorker: GIVE UP: %s\x1b[0m\n", job.TaskName)
				panic("give up, error: " + e.Error())
			}

			delay := e.Delay
			if nextJob != nil && nextJob.Retries == 0 {
				delay, _ = time.ParseDuration(nextJob.Interval)
			}

			interval := time.Duration(job.Retries) * delay
			taskIndex.activeInterval = time.NewTicker(interval)
			workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)

			trace.SetError(err)
			tags["is_retry"] = true
			tags["next_retry"] = time.Now().Add(interval).Format(time.RFC3339)

			job.Interval = interval.String()
			queue.PushJob(job)
		default:
			panic(e)
		}
	}
}
