package main

const (
	configsTemplate = `// {{.Header}}

package configs

import (
	"context"

	` + "{{ if .IsMonorepo }}\"monorepo/sdk\"\n{{end}}" + `"{{.PackagePrefix}}/pkg/shared"
	"{{.PackagePrefix}}/pkg/shared/repository"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/config"
	"{{.LibraryName}}/config/broker"
	{{ if not (or .SQLDeps .MongoDeps .RedisDeps) }}// {{ end }}"{{.LibraryName}}/config/database"
	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/middleware"
	"{{.LibraryName}}/validator"
)

// LoadConfigs load selected dependency configuration in this service
func LoadConfigs(baseCfg *config.Config) (deps dependency.Dependency) {

	var sharedEnv shared.Environment
	candihelper.MustParseEnv(&sharedEnv)
	shared.SetEnv(sharedEnv)

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		brokerDeps := broker.InitBrokers(
			{{if not .KafkaHandler}} // {{ end }}broker.SetKafka(broker.NewKafkaBroker()),
			{{if not .RabbitMQHandler}} // {{ end }}broker.SetRabbitMQ(broker.NewRabbitMQBroker()),
		)
		{{if not .RedisDeps}}// {{end}}redisDeps := database.InitRedis()
		{{if not .SQLDeps}}// {{end}}sqlDeps := database.InitSQLDatabase()
		{{if not .MongoDeps}}// {{end}}mongoDeps := database.InitMongoDB(ctx)
` + "{{ if .IsMonorepo }}\n		sdk.SetGlobalSDK(\n			// init service client sdk\n		)\n{{end}}" + `
		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(&shared.DefaultTokenValidator{}, &shared.DefaultACLPermissionChecker{})),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBroker(brokerDeps),
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
)
