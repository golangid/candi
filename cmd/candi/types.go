package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"text/template"
)

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
		"1": initService, "2": addModule, "3": addHandler, "4": initMonorepo, "5": runServiceMonorepo,
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

	tpl    *template.Template
	logger *log.Logger
	reader *bufio.Reader
)

type flagParameter struct {
	scopeFlag, packagePrefixFlag, protoOutputPkgFlag, outputFlag, libraryNameFlag string
	withGoModFlag                                                                 bool
	run, all                                                                      bool
	initService, addModule, addHandler, initMonorepo, version, isMonorepo         bool
	serviceName, monorepoProjectName                                              string
}

func (f *flagParameter) parseMonorepoFlag() error {
	f.packagePrefixFlag = "monorepo/services"
	f.withGoModFlag = false
	f.protoOutputPkgFlag = "monorepo/sdk"
	f.outputFlag = "services/"

	if (f.scopeFlag == "2" || f.scopeFlag == "3") && f.serviceName == "" {
		return fmt.Errorf(redFormat, "missing service name, make sure to include '-service' flag")
	}
	return nil
}

func (f *flagParameter) validateServiceName() (err error) {
	_, err = os.Stat(f.outputFlag + f.serviceName)
	if os.IsNotExist(err) {
		fmt.Printf(redFormat, fmt.Sprintf(`Service "%s" is not exist in "%s" directory`, f.serviceName, f.outputFlag))
		os.Exit(1)
	}
	return
}

func (f *flagParameter) validateModuleName(moduleName string) (err error) {
	_, err = os.Stat(f.outputFlag + f.serviceName + "/internal/modules/" + moduleName)
	if os.IsNotExist(err) {
		fmt.Printf(redFormat, fmt.Sprintf(`Module "%s" is not exist in service "%s"`, moduleName, f.serviceName))
		os.Exit(1)
	}
	return
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
