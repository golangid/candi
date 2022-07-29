package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func parseInput(flagParam *flagParameter) (srvConfig serviceConfig) {

	defer func() { srvConfig.parseDefaultHeader() }()

	serviceHandlers := make(map[string]bool)
	workerHandlers := make(map[string]bool)
	dependencies := make(map[string]bool)
	var cmdInput string
	var deliveryHandlerOption string
	var deliveryHandlerMap map[string]string

	scope, ok := scopeMap[flagParam.scopeFlag]
	switch scope {
	case InitService:
		srvConfig.ServiceName = inputServiceName()

	case AddModule:
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
				fmt.Printf(RedFormat, errMessage+", try again")
				goto stageInputServiceNameModule
			}
		}

	case AddHandler:
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
		if err := validateDir(flagParam.outputFlag + flagParam.serviceName + "/internal/modules/" + moduleName); scope != InitService && err == nil {
			fmt.Printf(RedFormat, "module '"+moduleName+"' is exist")
			goto stageInputModules
		}
		srvConfig.Modules = append(srvConfig.Modules, moduleConfig{
			ModuleName: strings.TrimSpace(candihelper.ToDelimited(moduleName, '-')), Skip: false,
		})
		flagParam.modules = append(flagParam.modules, strings.TrimSpace(moduleName))
	}
	if len(srvConfig.Modules) == 0 {
		fmt.Printf(RedFormat, "Modules cannot empty")
		goto stageInputModules
	}
	sort.Strings(flagParam.modules)

	if scope == AddModule {
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
			fmt.Printf(RedFormat, "Invalid option, try again")
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
			fmt.Printf(RedFormat, "Invalid option, try again")
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
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectWorkerHandlers
		}
	}

	if len(serviceHandlers) == 0 && len(workerHandlers) == 0 {
		fmt.Printf(RedFormat, "No server/worker handler selected, try again")
		goto stageSelectServerHandler
	}

	if scope == AddModule || scope == AddHandler {
		if b, ok := serviceHandlers[RestHandler]; ok {
			srvConfig.RestHandler = b
		}
		if b, ok := serviceHandlers[GrpcHandler]; ok {
			srvConfig.GRPCHandler = b
		}
		if b, ok := serviceHandlers[GraphqlHandler]; ok {
			srvConfig.GraphQLHandler = b
		}
		if b, ok := workerHandlers[KafkaHandler]; ok {
			srvConfig.KafkaHandler = b
		}
		if b, ok := workerHandlers[SchedulerHandler]; ok {
			srvConfig.SchedulerHandler = b
		}
		if b, ok := workerHandlers[RedissubsHandler]; ok {
			srvConfig.RedisSubsHandler = b
		}
		if b, ok := workerHandlers[TaskqueueHandler]; ok {
			srvConfig.TaskQueueHandler = b
		}
		if b, ok := workerHandlers[PostgresListenerHandler]; ok {
			srvConfig.PostgresListenerHandler = b
		}
		if b, ok := workerHandlers[RabbitmqHandler]; ok {
			srvConfig.RabbitMQHandler = b
		}
		srvConfig.checkWorkerActive()

		srvConfig.OutputDir = flagParam.outputFlag
		if scope == AddHandler {
			srvConfig.parseDefaultHeader()
			scopeAddHandler(flagParam, srvConfig, serviceHandlers, workerHandlers)
		}
		return
	}

stageSelectDependencies:
	cmdInput = readInput("Please select dependencies (separated by comma, enter for skip)\n" +
		"1) Redis\n" +
		"2) SQL Database\n" +
		"3) Mongo Database\n" +
		"4) Arango Database (plugin)")
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		str = strings.TrimSpace(str)
		if depsName, ok := dependencyMap[str]; ok {
			dependencies[depsName] = true
		} else if str != "" {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectDependencies
		}
	}
	if workerHandlers[RedissubsHandler] && !dependencies[RedisDeps] {
		fmt.Printf(RedFormat, "Redis Subscriber need redis, try again")
		goto stageSelectDependencies
	}
	if workerHandlers[TaskqueueHandler] && !dependencies[RedisDeps] {
		fmt.Printf(RedFormat, "Task Queue Worker need redis (for queue), try again")
		goto stageSelectDependencies
	}

	if dependencies[SqldbDeps] {
	stageSelectSQLDriver:
		cmdInput = readInput("Please select SQL database driver (choose one)\n" +
			"1) Postgres\n" +
			"2) MySQL")
		srvConfig.SQLDriver, ok = sqlDrivers[cmdInput]
		if !ok {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectSQLDriver
		}

	stageUseGORMLabel:
		cmdInput = readInput("Use GORM? (y/n)")
		if srvConfig.SQLUseGORM, ok = optionYesNo[cmdInput]; !ok {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageUseGORMLabel
		}
	}

	if workerHandlers[PostgresListenerHandler] && (!dependencies[SqldbDeps] || srvConfig.SQLDriver != "postgres") {
		fmt.Printf(RedFormat, "Postgres Event Listener Worker need Postgres config, try again")
		goto stageSelectDependencies
	}

stageUseLicense:
	cmdInput = readInput("Use License? (y/n)")
	isUsingLicense, ok := optionYesNo[cmdInput]
	if !ok {
		fmt.Printf(RedFormat, "Invalid option, try again")
		goto stageUseLicense
	}

	if isUsingLicense {
		srvConfig.Owner = inputOwnerName()
	stageSelectLicense:
		cmdInput = readInput("Please select your Product License (choose one)\n" +
			"1) MIT License\n" +
			"2) Apache License\n" +
			"3) Private License (if your product repository is private)")
		srvConfig.License, ok = licenseMap[cmdInput]
		if !ok {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectLicense
		}
	}

	// custom package name
	if packageOptions := strings.Split(os.Getenv(CandiPackagesEnv), ","); len(packageOptions) > 1 {
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
			fmt.Printf(RedFormat, "Invalid option, try again")
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

	srvConfig.LibraryName = flagParam.libraryNameFlag
	srvConfig.IsMonorepo = flagParam.isMonorepo
	srvConfig.RestHandler = serviceHandlers[RestHandler]
	srvConfig.GRPCHandler = serviceHandlers[GrpcHandler]
	srvConfig.GraphQLHandler = serviceHandlers[GraphqlHandler]
	srvConfig.KafkaHandler = workerHandlers[KafkaHandler]
	srvConfig.SchedulerHandler = workerHandlers[SchedulerHandler]
	srvConfig.RedisSubsHandler = workerHandlers[RedissubsHandler]
	srvConfig.TaskQueueHandler = workerHandlers[TaskqueueHandler]
	srvConfig.PostgresListenerHandler = workerHandlers[PostgresListenerHandler]
	srvConfig.RabbitMQHandler = workerHandlers[RabbitmqHandler]
	srvConfig.RedisDeps = dependencies[RedisDeps]
	srvConfig.SQLDeps, srvConfig.MongoDeps, srvConfig.ArangoDeps = dependencies[SqldbDeps], dependencies[MongodbDeps], dependencies[ArangodbDeps]
	srvConfig.checkWorkerActive()

	return
}
