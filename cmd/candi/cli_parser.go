package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/golangid/candi"
	"github.com/golangid/candi/candihelper"
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
		flagParam.moduleName = candihelper.ToDelimited(readInput("Please input existing module name to be added delivery handler(s):"), '-')
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
			ModuleName: strings.TrimSpace(candihelper.ToDelimited(moduleName, '-')), Skip: false,
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

	if scope == addModule || scope == addHandler {
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
		if scope == addHandler {
			scopeAddHandler(flagParam, srvConfig, serviceHandlers, workerHandlers)
		}
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
	srvConfig.GoVersion = getGoVersion()
	srvConfig.LibraryName = flagParam.libraryNameFlag

	// custom package name
	if packageOptions := strings.Split(os.Getenv(candiPackagesEnv), ","); len(packageOptions) > 1 {
	stageSelectPackageName:
		cliWording := "Please select package name (choose one)\n"
		inputPackageName := make(map[string]string, len(packageOptions))
		for i, packageOpt := range packageOptions {
			inputPackageName[strconv.Itoa(i+1)] = packageOpt
			cliWording += strconv.Itoa(i+1) + ") " + packageOpt + "\n"
		}
		cmdInput = readInput(strings.TrimSpace(cliWording))
		srvConfig.LibraryName, ok = inputPackageName[cmdInput]
		if !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			goto stageSelectPackageName
		}
	}

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
