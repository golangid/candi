package appfactory

import (
	redisworker "github.com/golangid/candi/codebase/app/redis_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRedisWorker setup redis worker with default config
func SetupRedisWorker(service factory.ServiceFactory, opts ...redisworker.OptionFunc) factory.AppServerFactory {
	redisOpts := []redisworker.OptionFunc{
		redisworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		redisworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	redisOpts = append(redisOpts, opts...)
	return redisworker.NewWorker(service, redisOpts...)
}
