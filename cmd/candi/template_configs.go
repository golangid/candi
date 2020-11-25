package main

const (
	configsTemplate = `// {{.Header}}

package configs

import (
	"context"

	"{{.GoModName}}/pkg/shared"
	"{{.GoModName}}/pkg/shared/repository"
	"{{$.GoModName}}/pkg/shared/usecase"

	"{{.PackageName}}/codebase/factory/dependency"
	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/config"
	{{if not .KafkaDeps}}// {{end}}"{{.PackageName}}/config/broker"
	{{ if not (or .SQLDeps .MongoDeps) }}// {{ end }}"{{.PackageName}}/config/database"
	"{{.PackageName}}/middleware"
	"{{.PackageName}}/validator"
)

// LoadConfigs load selected dependency configuration in this service
func LoadConfigs(baseCfg *config.Config) (deps dependency.Dependency) {

	loadAdditionalEnv()

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		{{if not .KafkaDeps}}// {{end}}kafkaDeps := broker.InitKafkaBroker()
		{{if not .RedisDeps}}// {{end}}redisDeps := database.InitRedis()
		{{if not .SQLDeps}}// {{end}}sqlDeps := database.InitSQLDatabase()
		{{if not .MongoDeps}}// {{end}}mongoDeps := database.InitMongoDB(ctx)

		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(&shared.DefaultTokenValidator{})),
			dependency.SetValidator(validator.NewValidator()),
			{{if not .KafkaDeps}}// {{end}}dependency.SetBroker(kafkaDeps),
			{{if not .RedisDeps}}// {{end}}dependency.SetRedisPool(redisDeps),
			{{if not .SQLDeps}}// {{end}}dependency.SetSQLDatabase(sqlDeps),
			{{if not .MongoDeps}}// {{end}}dependency.SetMongoDatabase(mongoDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{ // throw back to base config for close connection when application shutdown
			{{if not .KafkaDeps}}// {{end}}kafkaDeps,
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

package configs

// Environment additional in this service
type Environment struct {
	// more additional environment
}

var localEnv Environment

// GetEnv get global additional environment
func GetEnv() Environment {
	return localEnv
}

func loadAdditionalEnv() {
}
`
)
