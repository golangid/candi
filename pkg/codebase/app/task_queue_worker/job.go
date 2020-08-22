package taskqueueworker

import (
	"errors"
	"reflect"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Job model
type Job struct {
	ID           string                  `json:"id"`
	HandlerFunc  types.WorkerHandlerFunc `json:"-"`
	Args         []byte                  `json:"args"`
	WorkerIndex  int                     `json:"worker_index"`
	Retries      int                     `json:"retries"`
	MaxRetry     int                     `json:"max_retry"`
	interval     *time.Ticker
	nextInterval *time.Duration
}

// AddJob public function
func AddJob(jobID string, maxRetry int, args interface{}) error {

	jobFunc, ok := registeredJob[jobID]
	if !ok {
		return errors.New("job unregistered")
	}

	var newJob Job
	newJob.ID = jobID
	newJob.HandlerFunc = jobFunc
	newJob.Args = helper.ToBytes(args)
	newJob.MaxRetry = maxRetry
	newJob.interval = time.NewTicker(time.Second)

	registerJobInWorker(&newJob)

	return nil
}

func registerJobInWorker(job *Job) {
	mutex.Lock()
	defer mutex.Unlock()

	job.WorkerIndex = workerIndexJob[job.ID]
	workers[job.WorkerIndex].Chan = reflect.ValueOf(job.interval.C)
	queue := taskQueue[job.WorkerIndex]
	if queue == nil {
		queue = shared.NewQueue()
		taskQueue[job.WorkerIndex] = queue
	}
	queue.Push(job)

	refreshWorkerNotif <- struct{}{}
}

func getTaskQueue(workerIndex int) *Job {
	mutex.Lock()
	defer mutex.Unlock()

	queue := taskQueue[workerIndex]
	job := queue.Pop().(*Job)
	job.interval.Stop()

	func() {
		defer func() { recover() }()
		nextJob := queue.Peek()
		workers[workerIndex].Chan = reflect.ValueOf(nextJob.(*Job).interval.C)
		refreshWorkerNotif <- struct{}{}
	}()
	return job
}
