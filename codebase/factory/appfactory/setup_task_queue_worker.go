package appfactory

import (
	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupTaskQueueWorker setup task queue worker with default config
func SetupTaskQueueWorker(service factory.ServiceFactory, opts ...taskqueueworker.OptionFunc) factory.AppServerFactory {

	workerOpts := []taskqueueworker.OptionFunc{
		taskqueueworker.SetDashboardHTTPPort(env.BaseEnv().TaskQueueDashboardPort),
		taskqueueworker.SetMaxClientSubscriber(env.BaseEnv().TaskQueueDashboardMaxClientSubscribers),
		taskqueueworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	workerOpts = append(workerOpts, opts...)
	return taskqueueworker.NewTaskQueueWorker(service, workerOpts...)
}
