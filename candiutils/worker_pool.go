package candiutils

import (
	"context"
	"sync"
)

// WorkerPool implementation
type WorkerPool[T any] interface {
	Dispatch(ctx context.Context, jobFunc func(context.Context, T))
	AddJob(job T)
	Finish()
}

type workerPool[T any] struct {
	maxWorker int
	wg        sync.WaitGroup
	jobChan   chan T
}

// NewWorkerPool create an instance of WorkerPool.
func NewWorkerPool[T any](maxWorker int) WorkerPool[T] {
	wp := &workerPool[T]{
		maxWorker: maxWorker,
		wg:        sync.WaitGroup{},
		jobChan:   make(chan T),
	}

	return wp
}

func (wp *workerPool[T]) Dispatch(ctx context.Context, jobFunc func(context.Context, T)) {
	for i := 0; i < wp.maxWorker; i++ {
		go func(jobFunc func(context.Context, T)) {
			for job := range wp.jobChan {
				jobFunc(ctx, job)
				wp.wg.Done()
			}
		}(jobFunc)
	}
}

func (wp *workerPool[T]) AddJob(job T) {
	wp.wg.Add(1)
	wp.jobChan <- job
}

func (wp *workerPool[T]) Finish() {
	close(wp.jobChan)
	wp.wg.Wait()
}
