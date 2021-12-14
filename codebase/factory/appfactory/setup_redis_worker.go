package appfactory

import (
	redisworker "github.com/golangid/candi/codebase/app/redis_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRedisWorker setup cron worker with default config
func SetupRedisWorker(service factory.ServiceFactory) factory.AppServerFactory {
	redisOptions := []redisworker.OptionFunc{
		redisworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		redisworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	return redisworker.NewWorker(service, redisOptions...)
}
