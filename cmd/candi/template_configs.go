package main

const (
	configsTemplate = `// {{.Header}}

package configs

import (
	"context"

	{{ if not (or .SQLDeps .MongoDeps) }}// {{ end }}"{{.GoModName}}/pkg/shared/repository"

	{{ if not (or .SQLDeps .MongoDeps) }}// {{ end }}"{{.PackageName}}/candihelper"
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
		{{if not .KafkaDeps}}// {{end}}KafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		{{if not .RedisDeps}}// {{end}}redisDeps := database.InitRedis()
		{{if not .SQLDeps}}// {{end}}sqlDeps := database.InitSQLDatabase()
		{{if not .MongoDeps}}// {{end}}mongoDeps := database.InitMongoDB(ctx)

		extendedDeps := map[string]interface{}{
			{{if not .SQLDeps}}// {{end}}candihelper.RepositorySQL: repository.NewRepositorySQL(sqlDeps.ReadDB(), sqlDeps.WriteDB(), nil),
			{{if not .MongoDeps}}// {{end}}candihelper.RepositoryMongo: repository.NewRepositoryMongo(mongoDeps.ReadDB(), mongoDeps.WriteDB()),
		}

		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(nil)),
			dependency.SetValidator(validator.NewValidator()),
			{{if not .KafkaDeps}}// {{end}}dependency.SetBroker(KafkaDeps),
			{{if not .RedisDeps}}// {{end}}dependency.SetRedisPool(redisDeps),
			{{if not .SQLDeps}}// {{end}}dependency.SetSQLDatabase(sqlDeps),
			{{if not .MongoDeps}}// {{end}}dependency.SetMongoDatabase(mongoDeps),
			dependency.SetExtended(extendedDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{ // throw back to base config for close connection when application shutdown
			{{if not .KafkaDeps}}// {{end}}KafkaDeps,
			{{if not .RedisDeps}}// {{end}}redisDeps,
			{{if not .SQLDeps}}// {{end}}sqlDeps,
			{{if not .MongoDeps}}// {{end}}mongoDeps,
		}
	})

	return deps
}
`

	additionalEnvTemplate = `// {{.Header}}

package configs

// Environment additional in this service
type Environment struct {
	// more additional environment
}

var env Environment

// GetEnv get global additional environment
func GetEnv() Environment {
	return env
}

func loadAdditionalEnv() {
}
`
)
