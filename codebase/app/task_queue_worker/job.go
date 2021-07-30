package taskqueueworker

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultInterval = "1s"
)

// Job model
type Job struct {
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
	newJob.Arguments = string(args)
	newJob.MaxRetry = maxRetry
	newJob.Interval = defaultInterval
	newJob.Status = string(statusQueueing)
	newJob.CreatedAt = time.Now()

	go func(job Job, workerIndex int) {
		queue.PushJob(&job)
		repo.saveJob(job)
		broadcastAllToSubscribers()
		registerJobToWorker(&job, workerIndex)
		refreshWorkerNotif <- struct{}{}
	}(newJob, task.workerIndex)

	return nil
}

func registerJobToWorker(job *Job, workerIndex int) {
	interval, _ := time.ParseDuration(job.Interval)
	taskIndex := workerIndexTask[workerIndex]
	taskIndex.activeInterval = time.NewTicker(interval)
	workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
}
