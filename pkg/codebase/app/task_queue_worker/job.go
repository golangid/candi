package taskqueueworker

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

// Job model
type Job struct {
	ID       string `json:"id"`
	Args     []byte `json:"args"`
	Retries  int    `json:"retries"`
	MaxRetry int    `json:"max_retry"`
	Interval string `json:"interval"`
}

// AddJob public function
func AddJob(jobID string, maxRetry int, args interface{}) error {

	task, ok := registeredTask[jobID]
	if !ok {
		return errors.New("job unregistered")
	}

	var newJob Job
	newJob.ID = jobID
	newJob.Args = helper.ToBytes(args)
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

	job := queue.PopJob(taskIndex.taskID)
	job.Retries++

	tags := trace.Tags()
	tags["job_id"] = job.ID
	tags["job_args"] = string(job.Args)
	tags["retries"] = job.Retries
	tags["max_retry"] = job.MaxRetry

	nextJob := queue.NextJob(taskIndex.taskID)
	if nextJob != nil {
		registerJobToWorker(nextJob, workerIndex)
	}

	if err := registeredTask[job.ID].handlerFunc(ctx, job.Args); err != nil {
		switch e := err.(type) {
		case *ErrorRetrier:
			if job.Retries >= job.MaxRetry {
				fmt.Printf("\x1b[31;1mTaskQueueWorker: GIVE UP: %s\x1b[0m\n", job.ID)
				panic("give up, error: " + e.Error())
			}

			trace.SetError(e)
			tags["is_retry"] = true

			delay := e.Delay
			if nextJob != nil && nextJob.Retries == 0 {
				delay, _ = time.ParseDuration(nextJob.Interval)
			}

			interval := time.Duration(job.Retries) * delay
			taskIndex.activeInterval = time.NewTicker(interval)
			workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)

			tags["next_retry"] = time.Now().Add(interval).Format(time.RFC3339)

			job.Interval = interval.String()
			queue.PushJob(job)
		default:
			panic(e)
		}
	}
}
