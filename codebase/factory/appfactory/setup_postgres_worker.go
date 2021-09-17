package appfactory

import (
	"fmt"
	"time"

	"github.com/golangid/candi/candiutils"
	postgresworker "github.com/golangid/candi/codebase/app/postgres_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

func setupPostgresWorker(service factory.ServiceFactory) factory.AppServerFactory {
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
	return postgresworker.NewWorker(service, postgresOptions...)
}
