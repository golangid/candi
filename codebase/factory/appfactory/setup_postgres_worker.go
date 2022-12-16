package appfactory

import (
	postgresworker "github.com/golangid/candi/codebase/app/postgres_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupPostgresWorker setup postgres worker with default config
func SetupPostgresWorker(service factory.ServiceFactory, opts ...postgresworker.OptionFunc) factory.AppServerFactory {
	postgresOptions := []postgresworker.OptionFunc{
		postgresworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	postgresOptions = append(postgresOptions, opts...)
	return postgresworker.NewWorker(service, postgresOptions...)
}
