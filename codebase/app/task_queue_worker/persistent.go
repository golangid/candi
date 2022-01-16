package taskqueueworker

import "context"

// Persistent abstraction
type Persistent interface {
	FindAllJob(ctx context.Context, filter Filter) (jobs []Job)
	FindJobByID(ctx context.Context, id string) (job *Job, err error)
	CountAllJob(ctx context.Context, filter Filter) int
	AggregateAllTaskJob(ctx context.Context, filter Filter) (result []TaskResolver)
	SaveJob(ctx context.Context, job *Job)
	UpdateAllStatus(ctx context.Context, taskName string, currentStatus []JobStatusEnum, updatedStatus JobStatusEnum)
	CleanJob(ctx context.Context, taskName string)
	DeleteJob(ctx context.Context, id string) error
}
