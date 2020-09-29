package main

const (
	configsTemplate = `package configs

import (
	"context"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config"
	{{isActive $.kafkaDeps}}"pkg.agungdwiprasetyo.com/candi/config/broker"
	{{isActive $.isDatabaseActive}}"pkg.agungdwiprasetyo.com/candi/config/database"
	"pkg.agungdwiprasetyo.com/candi/middleware"
	"pkg.agungdwiprasetyo.com/candi/validator"
)

// LoadConfigs load selected dependency configuration in this service
func LoadConfigs(baseCfg *config.Config) (deps dependency.Dependency) {

	loadAdditionalEnv()

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		{{isActive $.kafkaDeps}}kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		{{isActive $.redisDeps}}redisDeps := database.InitRedis()
		{{isActive $.sqldbDeps}}sqlDeps := database.InitSQLDatabase()
		{{isActive $.mongodbDeps}}mongoDeps := database.InitMongoDB(ctx)

		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(nil)),
			dependency.SetValidator(validator.NewValidator()),
			{{isActive $.kafkaDeps}}dependency.SetBroker(kafkaDeps),
			{{isActive $.redisDeps}}dependency.SetRedisPool(redisDeps),
			{{isActive $.sqldbDeps}}dependency.SetSQLDatabase(sqlDeps),
			{{isActive $.mongodbDeps}}dependency.SetMongoDatabase(mongoDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{ // throw back to base config for close connection when application shutdown
			{{isActive $.kafkaDeps}}kafkaDeps,
			{{isActive $.redisDeps}}redisDeps,
			{{isActive $.sqldbDeps}}sqlDeps,
			{{isActive $.mongodbDeps}}mongoDeps,
		}
	})

	return deps
}
`

	additionalEnvTemplate = `package configs

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
