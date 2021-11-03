package main

const (
	ps1                      = "\x1b[32;1m>>> \x1b[0m"
	redFormat                = "\x1b[31;1m%s \x1b[0m\n"
	initService              = "initservice"
	addModule                = "addmodule"
	initMonorepo             = "initMonorepo"
	initMonorepoService      = "initMonorepoService"
	addModuleMonorepoService = "addModuleMonorepoService"
	runServiceMonorepo       = "runServiceMonorepo"
	addHandler               = "addHandler"

	restHandler             = "restHandler"
	grpcHandler             = "grpcHandler"
	graphqlHandler          = "graphqlHandler"
	kafkaHandler            = "kafkaHandler"
	schedulerHandler        = "schedulerHandler"
	redissubsHandler        = "redissubsHandler"
	taskqueueHandler        = "taskqueueHandler"
	postgresListenerHandler = "postgresListenerHandler"
	rabbitmqHandler         = "rabbitmqHandler"
	redisDeps               = "redisDeps"
	sqldbDeps               = "sqldbDeps"
	mongodbDeps             = "mongodbDeps"

	defaultPackageName = "github.com/golangid/candi"
	candiPackagesEnv   = "CANDI_CLI_PACKAGES"
)
