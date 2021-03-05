package main

import "text/template"

const (
	ps1                      = "\x1b[32;1m>>> \x1b[0m"
	redFormat                = "\x1b[31;1m%s \x1b[0m\n"
	initService              = "initservice"
	addModule                = "addmodule"
	initMonorepo             = "initMonorepo"
	initMonorepoService      = "initMonorepoService"
	addModuleMonorepoService = "addModuleMonorepoService"
	runServiceMonorepo       = "runServiceMonorepo"

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
	scopeMap = map[string]string{
		"1": initService, "2": addModule, "3": initMonorepo, "4": initMonorepoService, "5": addModuleMonorepoService, "6": runServiceMonorepo,
	}
	serviceHandlersMap = map[string]string{
		"1": restHandler, "2": grpcHandler, "3": graphqlHandler,
	}
	workerHandlersMap = map[string]string{
		"1": kafkaHandler, "2": schedulerHandler, "3": redissubsHandler, "4": taskqueueHandler,
	}
	dependencyMap = map[string]string{
		"1": redisDeps, "2": sqldbDeps, "3": mongodbDeps,
	}
	sqlDrivers = map[string]string{
		"1": "postgres", "2": "mysql",
	}

	tpl *template.Template
)

type flagParameter struct {
	scopeFlag, packagePrefixFlag, protoOutputPkgFlag, outputFlag, libraryNameFlag string
	withGoModFlag                                                                 bool
	run, all                                                                      bool
	initService, addModule, initMonorepo, version                                 bool
	serviceName                                                                   string
}

func (f *flagParameter) parseInitMonorepoService() {
	f.packagePrefixFlag = "monorepo/services"
	f.withGoModFlag = false
	f.protoOutputPkgFlag = "monorepo/sdk"
	f.outputFlag = "services/"
	f.scopeFlag = "4"
}
func (f *flagParameter) parseAddModuleMonorepoService() {
	f.packagePrefixFlag = "monorepo/services"
	f.withGoModFlag = false
	f.protoOutputPkgFlag = "monorepo/sdk"
	f.outputFlag = "services/"
	f.scopeFlag = "5"
}

type configHeader struct {
	Version       string
	Header        string
	LibraryName   string
	ServiceName   string
	PackagePrefix string
	ProtoSource   string
}

type config struct {
	RestHandler, GRPCHandler, GraphQLHandler                                           bool
	KafkaHandler, SchedulerHandler, RedisSubsHandler, TaskQueueHandler, IsWorkerActive bool
	RedisDeps, SQLDeps, MongoDeps, SQLUseGORM                                          bool
	SQLDriver                                                                          string
}

type serviceConfig struct {
	configHeader
	config
	Modules []moduleConfig
}

type moduleConfig struct {
	configHeader `json:"-"`
	config       `json:"-"`
	ModuleName   string
	Skip         bool `json:"-"`
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
	SkipFunc     func() bool
	Childs       []FileStructure
}
