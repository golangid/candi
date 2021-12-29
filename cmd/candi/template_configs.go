package main

const (
	configsTemplate = `// {{.Header}}

package configs

import (
	"context"

	` + "{{ if .IsMonorepo }}\"monorepo/sdk\"\n{{end}}" + `"{{.PackagePrefix}}/pkg/shared"
	"{{.PackagePrefix}}/pkg/shared/repository"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/broker"
	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/config"
	{{ if not (or .SQLDeps .MongoDeps .RedisDeps) }}// {{ end }}"{{.LibraryName}}/config/database"
	"{{.LibraryName}}/logger"
	"{{.LibraryName}}/middleware"
	"{{.LibraryName}}/tracer"
	"{{.LibraryName}}/validator"
)

// LoadServiceConfigs load selected dependency configuration in this service
func LoadServiceConfigs(baseCfg *config.Config) (deps dependency.Dependency) {

	var sharedEnv shared.Environment
	candihelper.MustParseEnv(&sharedEnv)
	shared.SetEnv(sharedEnv)

	logger.InitZap()
	tracer.InitOpenTracing(baseCfg.ServiceName)

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		brokerDeps := broker.InitBrokers(
			{{if not .KafkaHandler}}// {{ end }}broker.NewKafkaBroker(),
			{{if not .RabbitMQHandler}}// {{ end }}broker.NewRabbitMQBroker(),
		)
		{{if not .RedisDeps}}// {{end}}redisDeps := database.InitRedis()
		{{if not .SQLDeps}}// {{end}}sqlDeps := database.InitSQLDatabase()
		{{if not .MongoDeps}}// {{end}}mongoDeps := database.InitMongoDB(ctx)
` + "{{ if .IsMonorepo }}\n		sdk.SetGlobalSDK(\n			// init service client sdk\n		)\n{{end}}" + `
		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(
				&shared.DefaultTokenValidator{},
				&shared.DefaultACLPermissionChecker{}),
			),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBrokers(brokerDeps.GetBrokers()),
			{{if not .RedisDeps}}// {{end}}dependency.SetRedisPool(redisDeps),
			{{if not .SQLDeps}}// {{end}}dependency.SetSQLDatabase(sqlDeps),
			{{if not .MongoDeps}}// {{end}}dependency.SetMongoDatabase(mongoDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{ // throw back to base config for close connection when application shutdown
			brokerDeps,
			{{if not .RedisDeps}}// {{end}}redisDeps,
			{{if not .SQLDeps}}// {{end}}sqlDeps,
			{{if not .MongoDeps}}// {{end}}mongoDeps,
		}
	})

	repository.SetSharedRepository(deps)
	usecase.SetSharedUsecase(deps)

	return deps
}
`

	additionalEnvTemplate = `// {{.Header}}

package shared

// Environment additional in this service
type Environment struct {
	// more additional environment with struct tag is environment key example:
	// ExampleHost string ` + "`env:\"EXAMPLE_HOST\"`" + `
}

var sharedEnv Environment

// GetEnv get global additional environment
func GetEnv() Environment {
	return sharedEnv
}

// SetEnv get global additional environment
func SetEnv(env Environment) {
	sharedEnv = env
}
`

	appFactoryTemplate = `// {{.Header}}

package configs

import (
	"{{.LibraryName}}/codebase/factory"
	"{{.LibraryName}}/codebase/factory/appfactory"
	"{{.LibraryName}}/config/env"
)

/*
InitAppFromEnvironmentConfig constructor

Construct server/worker for running application from environment value

## Server
USE_REST=[bool]
USE_GRPC=[bool]
USE_GRAPHQL=[bool]

## Worker
USE_KAFKA_CONSUMER=[bool] # event driven handler
USE_CRON_SCHEDULER=[bool] # static scheduler
USE_REDIS_SUBSCRIBER=[bool] # dynamic scheduler
USE_TASK_QUEUE_WORKER=[bool]
USE_POSTGRES_LISTENER_WORKER=[bool]
USE_RABBITMQ_CONSUMER=[bool] # event driven handler and dynamic scheduler
*/
func InitAppFromEnvironmentConfig(service factory.ServiceFactory) (apps []factory.AppServerFactory) {

	if env.BaseEnv().UseKafkaConsumer {
		apps = append(apps, appfactory.SetupKafkaWorker(service))
	}
	if env.BaseEnv().UseCronScheduler {
		apps = append(apps, appfactory.SetupCronWorker(service))
	}
	if env.BaseEnv().UseTaskQueueWorker {
		apps = append(apps, appfactory.SetupTaskQueueWorker(service))
	}
	if env.BaseEnv().UseRedisSubscriber {
		apps = append(apps, appfactory.SetupRedisWorker(service))
	}
	if env.BaseEnv().UsePostgresListenerWorker {
		apps = append(apps, appfactory.SetupPostgresWorker(service))
	}
	if env.BaseEnv().UseRabbitMQWorker {
		apps = append(apps, appfactory.SetupRabbitMQWorker(service))
	}

	if env.BaseEnv().UseREST {
		apps = append(apps, appfactory.SetupRESTServer(service))
	}
	if env.BaseEnv().UseGRPC {
		apps = append(apps, appfactory.SetupGRPCServer(service))
	}
	if !env.BaseEnv().UseREST && env.BaseEnv().UseGraphQL {
		apps = append(apps, appfactory.SetupGraphQLServer(service))
	}

	return
}
`
)
