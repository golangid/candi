package taskqueueworker

import "context"

// QueueStorage abstraction for queue storage backend
type QueueStorage interface {
	PushJob(ctx context.Context, job *Job)
	PopJob(ctx context.Context, taskName string) (jobID string)
	NextJob(ctx context.Context, taskName string) (jobID string)
	Clear(ctx context.Context, taskName string)
}
