package taskqueueworker

// QueueStorage abstraction for queue storage backend
type QueueStorage interface {
	PushJob(job *Job)
	PopJob(taskName string) (jobID string)
	NextJob(taskName string) (jobID string)
	Clear(taskName string)
}
