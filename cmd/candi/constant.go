package main

import "strings"

const (
	Ps1                = "\x1b[32;1m>>> \x1b[0m"
	RedFormat          = "\x1b[31;1m%s \x1b[0m\n"
	InitService        = "initservice"
	AddModule          = "addmodule"
	InitMonorepo       = "initMonorepo"
	RunServiceMonorepo = "runServiceMonorepo"
	AddHandler         = "addHandler"
	AddUsecase         = "addUsecase"
	ApplyUsecase       = "applyUsecase"

	RestHandler             = "restHandler"
	GrpcHandler             = "grpcHandler"
	GraphqlHandler          = "graphqlHandler"
	KafkaHandler            = "kafkaHandler"
	SchedulerHandler        = "schedulerHandler"
	RedissubsHandler        = "redissubsHandler"
	TaskqueueHandler        = "taskqueueHandler"
	PostgresListenerHandler = "postgresListenerHandler"
	RabbitmqHandler         = "rabbitmqHandler"

	RedisDeps   = "redisDeps"
	SqldbDeps   = "sqldbDeps"
	MongodbDeps = "mongodbDeps"

	// plugin
	ArangodbDeps  = "arangodbDeps"
	FiberRestDeps = "fiberRestDeps"

	MitLicense     = "MIT License"
	ApacheLicense  = "Apache License"
	PrivateLicense = "Private License"

	DefaultPackageName = "github.com/golangid/candi"
	CandiPackagesEnv   = "CANDI_CLI_PACKAGES"
)

var deliveryHandlerLocation = map[string]string{
	RestHandler:             "resthandler/resthandler.go",
	GrpcHandler:             "grpchandler/grpchandler.go",
	GraphqlHandler:          "graphqlhandler/query_resolver.go",
	KafkaHandler:            "workerhandler/kafka_handler.go",
	SchedulerHandler:        "workerhandler/cron_handler.go",
	RedissubsHandler:        "workerhandler/redis_handler.go",
	TaskqueueHandler:        "workerhandler/taskqueue_handler.go",
	PostgresListenerHandler: "workerhandler/postgres_listener_handler.go",
	RabbitmqHandler:         "workerhandler/rabbitmq_handler.go",
	pluginGCPPubSubWorker:   "workerhandler/" + strings.ToLower(pluginGCPPubSubWorker) + "_handler.go",
	pluginSTOMPWorker:       "workerhandler/" + strings.ToLower(pluginSTOMPWorker) + "_handler.go",
}
