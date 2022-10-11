package appfactory

import (
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
	cronOptions = append(cronOptions, opts...)
	return cronworker.NewWorker(service, cronOptions...)
}
