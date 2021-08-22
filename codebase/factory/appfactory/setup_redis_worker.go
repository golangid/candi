package appfactory

import (
	"fmt"
	"time"

	"pkg.agungdp.dev/candi/candiutils"
	redisworker "pkg.agungdp.dev/candi/codebase/app/redis_worker"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupRedisWorker(service factory.ServiceFactory) factory.AppServerFactory {
	redisOptions := []redisworker.OptionFunc{
		redisworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		redisworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	if env.BaseEnv().UseConsul {
		consul, err := candiutils.NewConsul(&candiutils.ConsulConfig{
			ConsulAgentHost:   env.BaseEnv().ConsulAgentHost,
			ConsulKey:         fmt.Sprintf("%s_redis_worker", service.Name()),
			LockRetryInterval: time.Second,
			MaxJobRebalance:   env.BaseEnv().ConsulMaxJobRebalance,
		})
		if err != nil {
			panic(err)
		}
		redisOptions = append(redisOptions, redisworker.SetConsul(consul))
	}
	return redisworker.NewWorker(service, redisOptions...)
}
