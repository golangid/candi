package appfactory

import (
	"fmt"
	"time"

	"pkg.agungdp.dev/candi/candiutils"
	cronworker "pkg.agungdp.dev/candi/codebase/app/cron_worker"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupCronWorker(service factory.ServiceFactory) factory.AppServerFactory {
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
	return cronworker.NewWorker(service, cronOptions...)
}
