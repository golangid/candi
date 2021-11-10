package appfactory

import (
	"net/url"

	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupTaskQueueWorker setup cron worker with default config
func SetupTaskQueueWorker(service factory.ServiceFactory) factory.AppServerFactory {
	if service.GetDependency().GetRedisPool() == nil {
		panic("Task queue worker require redis for queue")
	}
	if service.GetDependency().GetMongoDatabase() == nil {
		panic("Task queue worker require mongo for dashboard management")
	}
	queue := taskqueueworker.NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
	persistent := taskqueueworker.NewMongoPersistent(service.GetDependency().GetMongoDatabase().WriteDB())
	var tracingDashboard string
	if env.BaseEnv().JaegerTracingDashboard != "" {
		tracingDashboard = env.BaseEnv().JaegerTracingDashboard
	} else if urlTracerAgent, _ := url.Parse("//" + env.BaseEnv().JaegerTracingHost); urlTracerAgent != nil {
		tracingDashboard = urlTracerAgent.Hostname()
	}
	return taskqueueworker.NewTaskQueueWorker(service,
		queue, persistent,
		taskqueueworker.SetDashboardHTTPPort(env.BaseEnv().TaskQueueDashboardPort),
		taskqueueworker.SetMaxClientSubscriber(env.BaseEnv().TaskQueueDashboardMaxClientSubscribers),
		taskqueueworker.SetJaegerTracingDashboard(tracingDashboard),
		taskqueueworker.SetDebugMode(env.BaseEnv().DebugMode),
	)
}
