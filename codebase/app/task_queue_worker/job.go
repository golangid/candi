package taskqueueworker

import (
	"reflect"
	"time"
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

func (job *Job) toMap() map[string]interface{} {
	return map[string]interface{}{
		"task_name":   job.TaskName,
		"arguments":   job.Arguments,
		"retries":     job.Retries,
		"max_retry":   job.MaxRetry,
		"interval":    job.Interval,
		"finished_at": job.FinishedAt,
		"status":      job.Status,
		"error":       job.Error,
		"trace_id":    job.TraceID,
	}
}

func registerJobToWorker(job *Job, workerIndex int) {

	interval, err := time.ParseDuration(job.Interval)
	if err != nil || interval <= 0 {
		return
	}

	taskIndex := runningWorkerIndexTask[workerIndex]
	taskIndex.activeInterval = time.NewTicker(interval)
	workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
}

func stopAllJob() {
	for _, task := range runningWorkerIndexTask {
		if task != nil && task.activeInterval != nil {
			task.activeInterval.Stop()
		}
	}
}

func stopAllJobInTask(taskName string) {
	t, ok := registeredTask[taskName]
	if !ok {
		return
	}

	if task := runningWorkerIndexTask[t.workerIndex]; task != nil {
		if task.activeInterval != nil {
			task.activeInterval.Stop()
		}
		if task.cancel != nil {
			task.cancel()
		}
	}
}
