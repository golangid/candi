package appfactory

import (
	"fmt"
	"time"

	"github.com/golangid/candi/candiutils"
	cronworker "github.com/golangid/candi/codebase/app/cron_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupCronWorker setup cron worker with default config, copy this function if you want to construct with custom config
func SetupCronWorker(service factory.ServiceFactory, opts ...cronworker.OptionFunc) factory.AppServerFactory {
	cronOptions := []cronworker.OptionFunc{
		cronworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		cronworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	if env.BaseEnv().UseConsul {
		consul, err := candiutils.NewConsul(&candiutils.ConsulConfig{
			ConsulAgentHost:   env.BaseEnv().ConsulAgentHost,
			ConsulKey:         fmt.Sprintf("%s_cron_worker", service.Name()),
			LockRetryInterval: time.Second,
			MaxJobRebalance:   env.BaseEnv().ConsulMaxJobRebalance,
		})
		if err != nil {
			panic(err)
		}
		cronOptions = append(cronOptions, cronworker.SetConsul(consul))
	}
	cronOptions = append(cronOptions, opts...)
	return cronworker.NewWorker(service, cronOptions...)
}
