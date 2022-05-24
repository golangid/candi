package appfactory

import (
	redisworker "github.com/golangid/candi/codebase/app/redis_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRedisWorker setup cron worker with default config
func SetupRedisWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return redisworker.NewWorker(service,
		redisworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		redisworker.SetDebugMode(env.BaseEnv().DebugMode),
	)
}
