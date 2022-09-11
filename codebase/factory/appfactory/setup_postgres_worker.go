package appfactory

import (
	"fmt"
	"time"

	"github.com/golangid/candi/candiutils"
	postgresworker "github.com/golangid/candi/codebase/app/postgres_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupPostgresWorker setup postgres worker with default config
func SetupPostgresWorker(service factory.ServiceFactory, opts ...postgresworker.OptionFunc) factory.AppServerFactory {
	postgresOptions := []postgresworker.OptionFunc{
		postgresworker.SetPostgresDSN(env.BaseEnv().DbSQLWriteDSN),
		postgresworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		postgresworker.SetDebugMode(env.BaseEnv().DebugMode),
	}
	if env.BaseEnv().UseConsul {
		consul, err := candiutils.NewConsul(&candiutils.ConsulConfig{
			ConsulAgentHost:   env.BaseEnv().ConsulAgentHost,
			ConsulKey:         fmt.Sprintf("%s_postgres_event_listener", service.Name()),
			LockRetryInterval: time.Second,
			MaxJobRebalance:   env.BaseEnv().ConsulMaxJobRebalance,
		})
		if err != nil {
			panic(err)
		}
		postgresOptions = append(postgresOptions, postgresworker.SetConsul(consul))
	}
	postgresOptions = append(postgresOptions, opts...)
	return postgresworker.NewWorker(service, postgresOptions...)
}
