package main

const (
	configsTemplate = `package configs

import (
	"context"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/config/broker"
	"pkg.agungdwiprasetyo.com/candi/config/database"
	"pkg.agungdwiprasetyo.com/candi/middleware"
	"pkg.agungdwiprasetyo.com/candi/validator"
)

// LoadConfigs load selected dependency configuration in this service
func LoadConfigs(baseCfg *config.Config) (deps dependency.Dependency) {

	loadAdditionalEnv()

	baseCfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		redisDeps := database.InitRedis()
		mongoDeps := database.InitMongoDB(ctx)

		// inject all service dependencies
		// See all option in dependency package
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(nil)),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBroker(kafkaDeps),
			dependency.SetRedisPool(redisDeps),
			dependency.SetMongoDatabase(mongoDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{kafkaDeps, redisDeps, mongoDeps} // throw back to base config for close connection when application shutdown
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
