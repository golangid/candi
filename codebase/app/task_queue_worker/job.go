package taskqueueworker

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/golangid/candi/candihelper"
)

type (
	// Job model
	Job struct {
		ID             string         `bson:"_id" json:"_id"`
		TaskName       string         `bson:"task_name" json:"task_name"`
		Arguments      string         `bson:"arguments" json:"arguments"`
		Retries        int            `bson:"retries" json:"retries"`
		MaxRetry       int            `bson:"max_retry" json:"max_retry"`
		Interval       string         `bson:"interval" json:"interval"`
		CreatedAt      time.Time      `bson:"created_at" json:"created_at"`
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
)

func (job *Job) updateValue() {
	if job.TraceID != "" && defaultOption.tracingDashboard != "" {
		job.TraceID = fmt.Sprintf("%s/%s", defaultOption.tracingDashboard, job.TraceID)
	}
	job.CreatedAt = job.CreatedAt.In(candihelper.AsiaJakartaLocalTime)
	if delay, err := time.ParseDuration(job.Interval); err == nil && job.Status == string(statusQueueing) {
		job.NextRetryAt = time.Now().Add(delay).In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
	}
	sort.Slice(job.RetryHistories, func(i, j int) bool {
		return job.RetryHistories[i].EndAt.After(job.RetryHistories[j].EndAt)
	})
	for i := range job.RetryHistories {
		job.RetryHistories[i].StartAt = job.RetryHistories[i].StartAt.In(candihelper.AsiaJakartaLocalTime)
		job.RetryHistories[i].EndAt = job.RetryHistories[i].EndAt.In(candihelper.AsiaJakartaLocalTime)

		if job.RetryHistories[i].TraceID != "" && defaultOption.tracingDashboard != "" {
			job.RetryHistories[i].TraceID = fmt.Sprintf("%s/%s", defaultOption.tracingDashboard, job.RetryHistories[i].TraceID)
		}
	}
}

func (job *Job) toMap() map[string]interface{} {
	return map[string]interface{}{
		"task_name":   job.TaskName,
		"arguments":   job.Arguments,
		"retries":     job.Retries,
		"max_retry":   job.MaxRetry,
		"interval":    job.Interval,
		"created_at":  job.CreatedAt,
		"finished_at": job.FinishedAt,
		"status":      job.Status,
		"error":       job.Error,
		"error_stack": job.ErrorStack,
		"trace_id":    job.TraceID,
	}
}

func registerJobToWorker(job *Job, workerIndex int) {
	// skip reinit ticker chan
	nextJob := queue.NextJob(context.Background(), job.TaskName)
	if len(semaphoreAddJob) > 1 && job.direct && nextJob != "" {
		return
	}

	interval, err := time.ParseDuration(job.Interval)
	if err != nil || interval <= 0 {
		return
	}

	taskIndex := runningWorkerIndexTask[workerIndex]
	if taskIndex.activeInterval == nil {
		taskIndex.activeInterval = time.NewTicker(interval)
	} else {
		taskIndex.activeInterval.Reset(interval)
	}
	workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
}

func stopAllJob() {
	for _, job := range runningWorkerIndexTask {
		if job != nil && job.activeInterval != nil {
			job.activeInterval.Stop()
		}
	}
}

func stopAllJobInTask(taskName string) {
	jobs, ok := registeredTask[taskName]
	if !ok {
		return
	}

	if job := runningWorkerIndexTask[jobs.workerIndex]; job != nil {
		if job.activeInterval != nil {
			job.activeInterval.Stop()
		}
		if job.cancel != nil {
			job.cancel()
		}
	}
}
