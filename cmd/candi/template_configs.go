package main

const (
	configsTemplate = `// {{.Header}}

package configs

import (
	"context"

	"{{.PackagePrefix}}/api"
	` + "{{ if .IsMonorepo }}\"monorepo/sdk\"\n	{{end}}" + `"{{.PackagePrefix}}/pkg/shared"
	"{{.PackagePrefix}}/pkg/shared/repository"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config"
	{{ if not (or .SQLDeps .MongoDeps .RedisDeps) }}// {{ end }}"github.com/golangid/candi/config/database"
	{{ if .ArangoDeps}} arango "github.com/golangid/candi-plugin/arangodb-adapter" {{ end }}
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/middleware"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/validator"
)

// LoadServiceConfigs load selected dependency configuration in this service
func LoadServiceConfigs(baseCfg *config.Config) (deps dependency.Dependency) {
	var sharedEnv shared.Environment
	candihelper.MustParseEnv(&sharedEnv)
	shared.SetEnv(sharedEnv)

	logger.InitZap()
	// logger.SetMaskLog(logger.NewMasker()) // add this for mask sensitive information

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		jaeger := tracer.InitJaeger(baseCfg.ServiceName)
		{{if not .RedisDeps}}// {{end}}redisDeps := database.InitRedis()
		{{if not .SQLDeps}}// {{end}}sqlDeps := database.InitSQLDatabase()
		{{if not .MongoDeps}}// {{end}}mongoDeps := database.InitMongoDB(ctx)` + `{{if .ArangoDeps}}
		arangoDeps := arango.InitArangoDB(ctx, sharedEnv.DbArangoReadHost, sharedEnv.DbArangoWriteHost){{end}}` + `
` + "{{ if .IsMonorepo }}\n		sdk.SetGlobalSDK(\n			// init service client sdk\n		)\n{{end}}" + `
		locker := {{if not .RedisDeps}}&candiutils.NoopLocker{}{{else}}candiutils.NewRedisLocker(redisDeps.WritePool()){{end}}

		brokerDeps := broker.InitBrokers(
			{{if not .KafkaHandler}}// {{ end }}broker.NewKafkaBroker(),
			{{if not .RabbitMQHandler}}// {{ end }}broker.NewRabbitMQBroker(),
			{{if not .RedisSubsHandler}}// {{ end }}broker.NewRedisBroker(redisDeps.WritePool()),
		)

		validatorDeps := validator.NewValidator()
		validatorDeps.JSONSchema.SchemaStorage = validator.NewFileSystemStorage(api.JSONSchema, "jsonschema")

		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetValidator(validatorDeps),
			dependency.SetBrokers(brokerDeps.GetBrokers()),
			dependency.SetLocker(locker),
			{{if not .RedisDeps}}// {{end}}dependency.SetRedisPool(redisDeps),
			{{if not .SQLDeps}}// {{end}}dependency.SetSQLDatabase(sqlDeps),
			{{if not .MongoDeps}}// {{end}}dependency.SetMongoDatabase(mongoDeps),{{if .ArangoDeps}}
			dependency.AddExtended("arangodb", arangoDeps),{{end}}
			// ... add more dependencies
		)
		return []interfaces.Closer{ // throw back to base config for close connection when application shutdown
			jaeger, deps,
		}
	})

	repository.SetSharedRepository(deps)
	usecase.SetSharedUsecase(deps)

	deps.SetMiddleware(middleware.NewMiddlewareWithOption(
		middleware.SetTokenValidator(&shared.DefaultMiddleware{}),
		middleware.SetACLPermissionChecker(&shared.DefaultMiddleware{}),
		middleware.SetUserIDExtractor(func(tokenClaim *candishared.TokenClaim) (userID string) {
			return tokenClaim.Subject
		}),
		{{if not .RedisDeps}}// {{end}}middleware.SetCache(deps.GetRedisPool().Cache(), middleware.DefaultCacheAge),
	))

	return deps
}
`

	additionalEnvTemplate = `// {{.Header}}

package shared

// Environment additional in this service
type Environment struct {
	// more additional environment with struct tag is environment key example:
	// ExampleHost string ` + "`env:\"EXAMPLE_HOST\"`" + `
	{{if .ArangoDeps}}DbArangoReadHost     	string	` + "`env:\"ARANGODB_HOST_READ\"`" + `
	DbArangoWriteHost      	string	` + "`env:\"ARANGODB_HOST_WRITE\"`" + `{{end}}
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
	"{{.PackagePrefix}}/api"

	"github.com/golangid/candi/candihelper"
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"{{if not .FiberRestHandler}}
	restserver "github.com/golangid/candi/codebase/app/rest_server"{{end}}
	{{if not .MongoDeps}}taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"{{end}}
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/appfactory"
	"github.com/golangid/candi/config/env"{{if .FiberRestHandler}}

	fiberrest "github.com/golangid/candi-plugin/fiber-rest"{{end}}{{if not (or .SQLDeps .MongoDeps)}}

	"database/sql"
	_ "github.com/mattn/go-sqlite3"{{end}}
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
		{{if .MongoDeps}}apps = append(apps, appfactory.SetupTaskQueueWorker(service)){{else if .SQLDeps}}persistent := taskqueueworker.NewSQLPersistent(service.
			GetDependency().GetSQLDatabase().WriteDB(),
		)
		apps = append(apps, appfactory.SetupTaskQueueWorker(service,
			taskqueueworker.SetPersistent(persistent),
		)){{else}}db, err := sql.Open("sqlite3", "./candi_task_queue_worker.db")
		if err != nil {
			panic(err)
		}
		persistent := taskqueueworker.NewSQLPersistent(db)
		apps = append(apps, appfactory.SetupTaskQueueWorker(service,
			taskqueueworker.SetPersistent(persistent),
		)){{end}}
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
		apps = append(apps, {{if .FiberRestHandler}}fiberrest.SetupFiberServer(service, fiberrest.AddGraphQLOption(
			graphqlserver.SetSchemaSource(candihelper.LoadAllFileFromFS(api.GraphQLSchema, ".", ".graphql")),
		)){{else}}appfactory.SetupRESTServer(service, restserver.AddGraphQLOption(
			graphqlserver.SetSchemaSource(candihelper.LoadAllFileFromFS(api.GraphQLSchema, ".", ".graphql")),
		)){{end}})
	} else if env.BaseEnv().UseGraphQL {
		apps = append(apps, appfactory.SetupGraphQLServer(service, graphqlserver.SetSchemaSource(
			candihelper.LoadAllFileFromFS(api.GraphQLSchema, ".", ".graphql"),
		)))
	}
	if env.BaseEnv().UseGRPC {
		apps = append(apps, appfactory.SetupGRPCServer(service))
	}

	return
}
`
)
