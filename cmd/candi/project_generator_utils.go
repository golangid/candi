package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"pkg.agungdp.dev/candi"
)

func parseInput(flagParam *flagParameter) (srvConfig serviceConfig) {

	serviceHandlers := make(map[string]bool)
	workerHandlers := make(map[string]bool)
	dependencies := make(map[string]bool)
	var cmdInput string
	var deliveryHandlerOption string
	var deliveryHandlerMap map[string]string

	scope, ok := scopeMap[flagParam.scopeFlag]
	switch scope {
	case initService:
		srvConfig.ServiceName = inputServiceName()

	case addModule:
		if flagParam.serviceName != "" {
			srvConfig.ServiceName = flagParam.serviceName
		} else if flagParam.isMonorepo {
		stageInputServiceNameModule:
			srvConfig.ServiceName = readInput("Please input existing service name to be added module(s):")
			_, err := os.Stat(flagParam.outputFlag + srvConfig.ServiceName)
			var errMessage string
			if strings.TrimSpace(srvConfig.ServiceName) == "" {
				errMessage = "Service name cannot empty"
			}
			if os.IsNotExist(err) {
				errMessage = fmt.Sprintf(`Service "%s" is not exist in "%s" directory`, srvConfig.ServiceName, flagParam.outputFlag)
			}
			if errMessage != "" {
				fmt.Printf(redFormat, errMessage+", try again")
				goto stageInputServiceNameModule
			}
		}

	case addHandler:
		flagParam.addHandler = true
		if flagParam.serviceName != "" {
			srvConfig.ServiceName = flagParam.serviceName
		} else if flagParam.isMonorepo {
		stageInputServiceName:
			flagParam.serviceName = readInput("Please input existing service name to be added delivery handler(s):")
			srvConfig.ServiceName = flagParam.serviceName
			if err := flagParam.validateServiceName(); err != nil {
				fmt.Print(err.Error())
				goto stageInputServiceName
			}
		}

	stageReadInputModule:
		flagParam.moduleName = readInput("Please input existing module name to be added delivery handler(s):")
		moduleDir := flagParam.getFullModuleChildDir()
		if err := validateDir(moduleDir); err != nil {
			fmt.Print(err.Error())
			goto stageReadInputModule
		}
		srvConfig = loadSavedConfig(flagParam)
		goto stageSelectServerHandler
	}

	flagParam.serviceName = srvConfig.ServiceName

stageInputModules:
	cmdInput = readInput("Please input new module names (if more than one, separated by comma):")
	for _, moduleName := range strings.Split(cmdInput, ",") {
		if err := validateDir(flagParam.outputFlag + flagParam.serviceName + "/internal/modules/" + moduleName); scope != initService && err == nil {
			fmt.Printf(redFormat, "module '"+moduleName+"' is exist")
			goto stageInputModules
		}
		srvConfig.Modules = append(srvConfig.Modules, moduleConfig{
			ModuleName: strings.TrimSpace(moduleName), Skip: false,
		})
		flagParam.modules = append(flagParam.modules, strings.TrimSpace(moduleName))
	}
	if len(srvConfig.Modules) == 0 {
		fmt.Printf(redFormat, "Modules cannot empty")
		goto stageInputModules
	}
	sort.Strings(flagParam.modules)

	if scope == addModule {
		savedConfig := loadSavedConfig(flagParam)
		savedConfig.Modules = append(savedConfig.Modules, srvConfig.Modules...)
		srvConfig = savedConfig

	stageChooseCustomConfig:
		format := "\n* REST: %t\n* GRPC: %t\n* GraphQL: %t\n* Kafka: %t\n* Scheduler: %t\n" +
			"* RedisSubs: %t\n* TaskQueue: %t\n* PostgresListener: %t\n* RabbitMQ: %t"
		currentConfig := fmt.Sprintf(format,
			savedConfig.RestHandler, savedConfig.GRPCHandler, savedConfig.GraphQLHandler, savedConfig.KafkaHandler, savedConfig.SchedulerHandler,
			savedConfig.RedisSubsHandler, savedConfig.TaskQueueHandler, savedConfig.PostgresListenerHandler, savedConfig.RabbitMQHandler)
		customConfig := readInput(fmt.Sprintf("Use custom server/worker handler (y/n)? \ncurrent handler config is: %s", currentConfig))
		yes, ok := optionYesNo[customConfig]
		if !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageChooseCustomConfig
		}

		if yes {
			srvConfig.disableAllHandler()
			goto stageSelectServerHandler
		}
		return
	}

stageSelectServerHandler:
	deliveryHandlerOption, deliveryHandlerMap = filterServerHandler(srvConfig, flagParam)
	if len(deliveryHandlerMap) == 0 {
		goto stageSelectWorkerHandlers
	}
	cmdInput = readInput("Please select server handlers (separated by comma, enter for skip)\n" + deliveryHandlerOption)
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if serverName, ok := deliveryHandlerMap[strings.TrimSpace(str)]; ok {
			serviceHandlers[serverName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageSelectServerHandler
		}
	}

stageSelectWorkerHandlers:
	deliveryHandlerOption, deliveryHandlerMap = filterWorkerHandler(srvConfig, flagParam)
	cmdInput = readInput("Please select worker handlers (separated by comma, enter for skip)\n" + deliveryHandlerOption)
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if workerName, ok := deliveryHandlerMap[strings.TrimSpace(str)]; ok {
			workerHandlers[workerName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageSelectWorkerHandlers
		}
	}

	if len(serviceHandlers) == 0 && len(workerHandlers) == 0 {
		fmt.Printf(redFormat, "No server/worker handler selected, try again")
		goto stageSelectServerHandler
	}

	if scope == addHandler {
		if b, ok := serviceHandlers[restHandler]; ok {
			srvConfig.RestHandler = b
		}
		if b, ok := serviceHandlers[grpcHandler]; ok {
			srvConfig.GRPCHandler = b
		}
		if b, ok := serviceHandlers[graphqlHandler]; ok {
			srvConfig.GraphQLHandler = b
		}
		if b, ok := workerHandlers[kafkaHandler]; ok {
			srvConfig.KafkaHandler = b
		}
		if b, ok := workerHandlers[schedulerHandler]; ok {
			srvConfig.SchedulerHandler = b
		}
		if b, ok := workerHandlers[redissubsHandler]; ok {
			srvConfig.RedisSubsHandler = b
		}
		if b, ok := workerHandlers[taskqueueHandler]; ok {
			srvConfig.TaskQueueHandler = b
		}
		if b, ok := workerHandlers[postgresListenerHandler]; ok {
			srvConfig.PostgresListenerHandler = b
		}
		if b, ok := workerHandlers[rabbitmqHandler]; ok {
			srvConfig.RabbitMQHandler = b
		}
		srvConfig.checkWorkerActive()

		srvConfig.OutputDir = flagParam.outputFlag
		scopeAddHandler(flagParam, srvConfig, serviceHandlers, workerHandlers)
		return
	}

stageSelectDependencies:
	cmdInput = readInput("Please select dependencies (separated by comma, enter for skip)\n" +
		"1) Redis\n" +
		"2) SQL Database\n" +
		"3) Mongo Database")
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		str = strings.TrimSpace(str)
		if depsName, ok := dependencyMap[str]; ok {
			dependencies[depsName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageSelectDependencies
		}
	}
	if workerHandlers[redissubsHandler] && !dependencies[redisDeps] {
		fmt.Printf(redFormat, "Redis Subscriber need redis, try again")
		goto stageSelectDependencies
	}
	if workerHandlers[taskqueueHandler] && !(dependencies[redisDeps] && dependencies[mongodbDeps]) {
		fmt.Printf(redFormat, "Task Queue Worker need redis (for queue) and mongo (for log storage), try again")
		goto stageSelectDependencies
	}

	if dependencies[sqldbDeps] {
	stageSelectSQLDriver:
		cmdInput = readInput("Please select SQL database driver (choose one)\n" +
			"1) Postgres\n" +
			"2) MySQL")
		srvConfig.SQLDriver, ok = sqlDrivers[cmdInput]
		if !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageSelectSQLDriver
		}

	stageUseGORMLabel:
		cmdInput = readInput("Use GORM? (y/n)")
		if srvConfig.SQLUseGORM, ok = optionYesNo[cmdInput]; !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageUseGORMLabel
		}
	}

	if workerHandlers[postgresListenerHandler] && (!dependencies[sqldbDeps] || srvConfig.SQLDriver != "postgres") {
		fmt.Printf(redFormat, "Postgres Event Listener Worker need Postgres config, try again")
		goto stageSelectDependencies
	}

	srvConfig.Header = fmt.Sprintf("Code generated by candi %s.", candi.Version)
	srvConfig.Version = candi.Version
	srvConfig.LibraryName = flagParam.libraryNameFlag
	if flagParam.packagePrefixFlag != "" {
		flagParam.packagePrefixFlag = strings.TrimSuffix(flagParam.packagePrefixFlag, "/") + "/"
		srvConfig.PackagePrefix = flagParam.packagePrefixFlag + srvConfig.ServiceName
	} else {
		srvConfig.PackagePrefix = srvConfig.ServiceName
	}
	if flagParam.protoOutputPkgFlag != "" {
		srvConfig.ProtoSource = flagParam.protoOutputPkgFlag + "/" + srvConfig.ServiceName + "/proto"
	} else {
		srvConfig.ProtoSource = srvConfig.PackagePrefix + "/api/proto"
	}

	srvConfig.IsMonorepo = flagParam.isMonorepo
	srvConfig.RestHandler = serviceHandlers[restHandler]
	srvConfig.GRPCHandler = serviceHandlers[grpcHandler]
	srvConfig.GraphQLHandler = serviceHandlers[graphqlHandler]
	srvConfig.KafkaHandler = workerHandlers[kafkaHandler]
	srvConfig.SchedulerHandler = workerHandlers[schedulerHandler]
	srvConfig.RedisSubsHandler = workerHandlers[redissubsHandler]
	srvConfig.TaskQueueHandler = workerHandlers[taskqueueHandler]
	srvConfig.PostgresListenerHandler = workerHandlers[postgresListenerHandler]
	srvConfig.RabbitMQHandler = workerHandlers[rabbitmqHandler]
	srvConfig.RedisDeps = dependencies[redisDeps]
	srvConfig.SQLDeps, srvConfig.MongoDeps = dependencies[sqldbDeps], dependencies[mongodbDeps]
	srvConfig.checkWorkerActive()

	return
}

func parseSharedRepository(data serviceConfig) (repos []FileStructure) {
	for i := range data.Modules {
		data.Modules[i].config = data.config
	}
	repos = append(repos, []FileStructure{
		{FromTemplate: true, DataSource: data, Source: templateRepository, FileName: "repository.go"},
		{FromTemplate: true, DataSource: data, Source: templateRepositoryUOWSQL, FileName: "repository_sql.go"},
		{FromTemplate: true, DataSource: data, Source: templateRepositoryUOWMongo, FileName: "repository_mongo.go"},
	}...)
	return
}

func parseRepositoryModule(data moduleConfig) (repos []FileStructure) {
	repos = append(repos, []FileStructure{
		{FromTemplate: true, DataSource: data, Source: templateRepositorySQLImpl, FileName: "repository_sql.go"},
		{FromTemplate: true, DataSource: data, Source: templateRepositoryMongoImpl, FileName: "repository_mongo.go"},
	}...)
	return
}

func loadTemplate(source string, sourceData interface{}) []byte {
	var byteBuff = new(bytes.Buffer)
	defer byteBuff.Reset()

	tmpl, err := tpl.Funcs(formatTemplate()).Parse(source)
	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(byteBuff, sourceData); err != nil {
		panic(err)
	}

	return byteBuff.Bytes()
}

func formatTemplate() template.FuncMap {
	return template.FuncMap{

		"clean": func(v string) string {
			return cleanSpecialChar.Replace(v)
		},
		"cleanPathModule": func(v string) string {
			return modulePathReplacer.Replace(v)
		},
		"upper": func(str string) string {
			return strings.Title(str)
		},
		"lower": func(str string) string {
			return strings.ToLower(str)
		},
		"isActive": func(str string) string {
			ok, _ := strconv.ParseBool(str)
			if ok {
				return ""
			}
			return "// "
		},
	}
}

func mergeMap(dest, source map[string]interface{}) {
	for k, v := range source {
		dest[k] = v
	}
}

func printBanner() {
	fmt.Printf(`
	 _____   ___   _   _______ _____ 
	/  __ \ / _ \ | \ | |  _  \_   _|
	| /  \// /_\ \|  \| | | | | | |  
	| |    |  _  || . | | | | | | |  
	| \__/\| | | || |\  | |/ / _| |_ 
	 \____/\_| |_/\_| \_/___/  \___/  %s

`, candi.Version)
}

func isWorkdirMonorepo() bool {
	_, errSdk := ioutil.ReadDir("sdk/")
	_, errService := ioutil.ReadDir("services/")
	return (errSdk == nil) && (errService == nil)
}

func inputServiceName() (serviceName string) {
	serviceName = readInput("Please input service name:")
	_, err := os.Stat(serviceName)
	var errMessage string
	if strings.TrimSpace(serviceName) == "" {
		errMessage = "Service name cannot empty"
	}
	if !os.IsNotExist(err) {
		errMessage = "Folder already exists"
	}
	if errMessage != "" {
		fmt.Printf(redFormat, errMessage+", try again")
		inputServiceName()
	}
	return
}

func readInput(cmd string) string {
	logger.Printf("\033[1m%s\033[0m ", cmd)
	fmt.Printf(">> ")
	cmdInput, _ := reader.ReadString('\n')
	return strings.TrimRight(strings.TrimSpace(cmdInput), "\n")
}

func validateDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf(redFormat, fmt.Sprintf(`Directory "%s" is not exist`, dir))
	}
	return nil
}

func isDirExist(dir string) bool {
	_, err := os.Stat(dir)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func loadSavedConfig(flagParam *flagParameter) serviceConfig {
	var baseDir string
	if flagParam.serviceName != "" {
		baseDir = flagParam.outputFlag + flagParam.serviceName + "/"
	}

	b, err := ioutil.ReadFile(baseDir + "candi.json")
	if err != nil {
		log.Fatal("ERROR: cannot find candi.json file")
	}
	var savedConfig serviceConfig
	json.Unmarshal(b, &savedConfig)
	for i := range savedConfig.Modules {
		savedConfig.Modules[i].Skip = true
	}
	return savedConfig
}

func filterServerHandler(cfg serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if !cfg.RestHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "resthandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) REST API", len(options)+1))
		handlers[strconv.Itoa(len(options))] = restHandler
	}
	if !cfg.GRPCHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "grpchandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) GRPC", len(options)+1))
		handlers[strconv.Itoa(len(options))] = grpcHandler
	}
	if !cfg.GraphQLHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "graphqlhandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) GraphQL", len(options)+1))
		handlers[strconv.Itoa(len(options))] = graphqlHandler
	}

	wording = strings.Join(options, "\n")
	return
}

func filterWorkerHandler(cfg serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if !cfg.KafkaHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "kafka_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Kafka Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = kafkaHandler
	}
	if !cfg.SchedulerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "cron_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Scheduler", len(options)+1))
		handlers[strconv.Itoa(len(options))] = schedulerHandler
	}
	if !cfg.RedisSubsHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "redis_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Redis Subscriber", len(options)+1))
		handlers[strconv.Itoa(len(options))] = redissubsHandler
	}
	if !cfg.TaskQueueHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "taskqueue_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Task Queue Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = taskqueueHandler
	}
	if !cfg.PostgresListenerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "postgres_listener_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Postgres Event Listener Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = postgresListenerHandler
	}
	if !cfg.RabbitMQHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "rabbitmq_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) RabbitMQ Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = rabbitmqHandler
	}

	wording = strings.Join(options, "\n")
	return
}

func readFileAndApply(filepath string, oldContent, newContent string) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	os.WriteFile(filepath, bytes.Replace(b, []byte(oldContent), []byte(newContent), -1), 0644)
}
