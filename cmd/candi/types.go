package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	scopeMap = map[string]string{
		"1": InitMonorepo, "2": InitService, "3": AddModule, "4": AddHandler, "5": RunServiceMonorepo,
	}
	serviceHandlersMap = map[string]string{
		"1": RestHandler, "2": GrpcHandler, "3": GraphqlHandler,
	}
	workerHandlersMap = map[string]string{
		"1": KafkaHandler, "2": SchedulerHandler, "3": RedissubsHandler, "4": TaskqueueHandler, "5": PostgresListenerHandler, "6": RabbitmqHandler,
	}
	dependencyMap = map[string]string{
		"1": RedisDeps, "2": SqldbDeps, "3": MongodbDeps, "4": ArangodbDeps,
	}
	sqlDrivers = map[string]string{
		"1": "postgres", "2": "mysql",
	}
	optionYesNo = map[string]bool{"y": true, "n": false}
	licenseMap  = map[string]string{
		"1": MitLicense, "2": ApacheLicense, "3": PrivateLicense,
	}
	licenseMapTemplate = map[string]string{
		MitLicense: mitLicenseTemplate, ApacheLicense: apacheLicenseTemplate, PrivateLicense: privateLicenseTemplate,
	}

	tpl    *template.Template
	logger *log.Logger
	reader *bufio.Reader

	specialChar        = []string{"*", "", "/", "", ":", ""}
	cleanSpecialChar   = strings.NewReplacer(append(specialChar, "-", "")...)
	modulePathReplacer = strings.NewReplacer(specialChar...)
)

type flagParameter struct {
	scopeFlag, packagePrefixFlag, protoOutputPkgFlag, outputFlag, libraryNameFlag string
	withGoModFlag                                                                 bool
	run, all                                                                      bool
	initService, addModule, addHandler, initMonorepo, version, isMonorepo         bool
	serviceName, moduleName, monorepoProjectName                                  string
	modules                                                                       []string
}

func (f *flagParameter) parseMonorepoFlag() error {
	f.packagePrefixFlag = "monorepo/services"
	f.withGoModFlag = false
	f.protoOutputPkgFlag = "monorepo/sdk"
	f.outputFlag = "services/"

	if (f.scopeFlag == "2" || f.scopeFlag == "3") && f.serviceName == "" {
		return fmt.Errorf(RedFormat, "missing service name, make sure to include '-service' flag")
	}
	return nil
}

func (f *flagParameter) validateServiceName() error {
	_, err := os.Stat(f.outputFlag + f.serviceName)
	if os.IsNotExist(err) {
		return fmt.Errorf(RedFormat, fmt.Sprintf(`Service "%s" is not exist in "%s" directory`, f.serviceName, f.outputFlag))
	}
	return nil
}

func (f *flagParameter) validateModuleName(moduleName string) (err error) {
	_, err = os.Stat(f.outputFlag + f.serviceName + "/internal/modules/" + moduleName)
	if os.IsNotExist(err) {
		fmt.Printf(RedFormat, fmt.Sprintf(`Module "%s" is not exist in service "%s"`, moduleName, f.serviceName))
		os.Exit(1)
	}
	return
}

func (f *flagParameter) getFullModuleChildDir(paths ...string) string {
	paths = append([]string{f.moduleName}, paths...)
	return strings.TrimPrefix(f.outputFlag+f.serviceName+"/internal/modules/"+strings.Join(paths, "/"), "/")
}

type configHeader struct {
	GoVersion     string
	Version       string
	Header        string `json:"-"`
	LibraryName   string
	ServiceName   string
	PackagePrefix string
	ProtoSource   string
	OutputDir     string `json:"-"`
	Owner         string
	License       string
	Year          int
}

type config struct {
	IsMonorepo                                                         bool
	RestHandler, GRPCHandler, GraphQLHandler                           bool
	KafkaHandler, SchedulerHandler, RedisSubsHandler, TaskQueueHandler bool
	PostgresListenerHandler, RabbitMQHandler, IsWorkerActive           bool
	RedisDeps, SQLDeps, MongoDeps, SQLUseGORM, ArangoDeps              bool
	SQLDriver                                                          string
}

type serviceConfig struct {
	configHeader
	config
	Modules []moduleConfig
}

func (s *serviceConfig) checkWorkerActive() {
	s.IsWorkerActive = s.KafkaHandler ||
		s.SchedulerHandler ||
		s.RedisSubsHandler ||
		s.PostgresListenerHandler ||
		s.TaskQueueHandler ||
		s.RabbitMQHandler
}
func (s *serviceConfig) disableAllHandler() {
	s.RestHandler = false
	s.GRPCHandler = false
	s.GraphQLHandler = false
	s.KafkaHandler = false
	s.SchedulerHandler = false
	s.RedisSubsHandler = false
	s.TaskQueueHandler = false
	s.PostgresListenerHandler = false
	s.RabbitMQHandler = false
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
	SkipFunc     func() bool `json:"-"`
	SkipIfExist  bool
	Childs       []FileStructure
}
