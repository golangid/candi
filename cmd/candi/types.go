package main

import "text/template"

const (
	ps1         = "\x1b[32;1m>>> \x1b[0m"
	packageName = "pkg.agungdwiprasetyo.com/candi"
	initService = "initservice"
	addModule   = "addmodule"
	version     = "v0.0.7"

	restHandler      = "restHandler"
	grpcHandler      = "grpcHandler"
	graphqlHandler   = "graphqlHandler"
	kafkaHandler     = "kafkaHandler"
	schedulerHandler = "schedulerHandler"
	redissubsHandler = "redissubsHandler"
	taskqueueHandler = "taskqueueHandler"
	kafkaDeps        = "kafkaDeps"
	redisDeps        = "redisDeps"
	sqldbDeps        = "sqldbDeps"
	mongodbDeps      = "mongodbDeps"
)

var (
	scopeFlag, serviceNameFlag, gomodName string
	scopeMap                              = map[string]string{
		"1": initService, "2": addModule,
	}
	serviceHandlersMap = map[string]string{
		"1": restHandler, "2": grpcHandler, "3": graphqlHandler,
	}
	workerHandlersMap = map[string]string{
		"1": kafkaHandler, "2": schedulerHandler, "3": redissubsHandler, "4": taskqueueHandler,
	}
	dependencyMap = map[string]string{
		"1": kafkaDeps, "2": redisDeps, "3": sqldbDeps, "4": mongodbDeps,
	}
	sqlDrivers = map[string]string{
		"1": "postgres", "2": "mysql",
	}

	tpl *template.Template
)

type configHeader struct {
	Header      string
	PackageName string
	ServiceName string
	GoModName   string
}

type config struct {
	RestHandler, GRPCHandler, GraphQLHandler                                           bool
	KafkaHandler, SchedulerHandler, RedisSubsHandler, TaskQueueHandler, IsWorkerActive bool
	KafkaDeps, RedisDeps, SQLDeps, MongoDeps                                           bool
	SQLDriver                                                                          string
}

type serviceConfig struct {
	configHeader
	config
	Modules []moduleConfig
}

type moduleConfig struct {
	configHeader
	config
	ModuleName string
	Skip       bool
}

// FileStructure model
type FileStructure struct {
	TargetDir    string
	IsDir        bool
	FromTemplate bool
	DataSource   interface{}
	Source       string
	FileName     string
	Skip         bool
	Childs       []FileStructure
}
