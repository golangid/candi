package appfactory

import (
	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupTaskQueueWorker setup cron worker with default config
func SetupTaskQueueWorker(service factory.ServiceFactory, opts ...taskqueueworker.OptionFunc) factory.AppServerFactory {
	if service.GetDependency().GetRedisPool() == nil {
		panic("Task queue worker require redis for queue")
	}
	queue := taskqueueworker.NewRedisQueue(service.
		GetDependency().
		GetRedisPool().
		WritePool(),
	)

	workerOpts := []taskqueueworker.OptionFunc{
		taskqueueworker.SetQueue(queue),
		taskqueueworker.SetDashboardHTTPPort(env.BaseEnv().TaskQueueDashboardPort),
		taskqueueworker.SetMaxClientSubscriber(env.BaseEnv().TaskQueueDashboardMaxClientSubscribers),
		taskqueueworker.SetTracingDashboard(env.BaseEnv().JaegerTracingDashboard + "/trace"),
		taskqueueworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	workerOpts = append(workerOpts, opts...)
	return taskqueueworker.NewTaskQueueWorker(service, workerOpts...)
}
