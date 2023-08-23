package main

const (
	Ps1                = "\x1b[32;1m>>> \x1b[0m"
	RedFormat          = "\x1b[31;1m%s \x1b[0m\n"
	InitService        = "initservice"
	AddModule          = "addmodule"
	InitMonorepo       = "initMonorepo"
	RunServiceMonorepo = "runServiceMonorepo"
	AddHandler         = "addHandler"

	RestHandler             = "restHandler"
	GrpcHandler             = "grpcHandler"
	GraphqlHandler          = "graphqlHandler"
	KafkaHandler            = "kafkaHandler"
	SchedulerHandler        = "schedulerHandler"
	RedissubsHandler        = "redissubsHandler"
	TaskqueueHandler        = "taskqueueHandler"
	PostgresListenerHandler = "postgresListenerHandler"
	RabbitmqHandler         = "rabbitmqHandler"
	RedisDeps               = "redisDeps"
	SqldbDeps               = "sqldbDeps"
	MongodbDeps             = "mongodbDeps"

	// plugin
	ArangodbDeps  = "arangodbDeps"
	FiberRestDeps = "fiberRestDeps"

	MitLicense     = "MIT License"
	ApacheLicense  = "Apache License"
	PrivateLicense = "Private License"

	DefaultPackageName = "github.com/golangid/candi"
	CandiPackagesEnv   = "CANDI_CLI_PACKAGES"
)
