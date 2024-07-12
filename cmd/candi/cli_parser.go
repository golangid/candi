package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func parseInput(flagParam *flagParameter) (srvConfig serviceConfig) {
	defer func() { srvConfig.parseDefaultHeader() }()

	handlers := make(map[string]bool)
	dependencies := make(map[string]bool)
	var cmdInput string
	var deliveryHandlerOption string
	var deliveryHandlerMap map[string]string
	var newModules []moduleConfig

	srvConfig = loadSavedConfig(flagParam)

	scope, ok := scopeMap[flagParam.scopeFlag]
	switch scope {
	case InitService:
		flagParam.serviceName = cliStageInputServiceName()
		srvConfig.ServiceName = flagParam.serviceName

	case AddModule:
		flagParam.addModule = true
		if flagParam.isMonorepo {
			flagParam.serviceName = cliStageInputExistingServiceName(flagParam.outputFlag,
				"Please input existing service name to be added module(s):")
			srvConfig = loadSavedConfig(flagParam)
		}

	case AddHandler:
		flagParam.addHandler = true
		if flagParam.isMonorepo {
			flagParam.serviceName = cliStageInputExistingServiceName(flagParam.outputFlag,
				"Please input existing service name to be added delivery handler(s):")
			srvConfig = loadSavedConfig(flagParam)
			srvConfig.ServiceName = flagParam.serviceName
		}

		flagParam.moduleName = cliStageInputModule(filepath.Join(flagParam.outputFlag, flagParam.serviceName, "internal/modules"),
			"Please input existing module name to be added delivery handler(s):")
		goto stageSelectServerHandler

	case AddUsecase:
		if flagParam.isMonorepo {
			flagParam.serviceName = cliStageInputExistingServiceName(flagParam.outputFlag,
				"Please input existing service name to be added usecase(s):")
		}
		flagParam.moduleName = cliStageInputModule(filepath.Join(flagParam.outputFlag, flagParam.serviceName, "internal/modules"),
			"Please input existing module name to be added usecase:")
		ucName := cliStageInputUsecaseName(flagParam.getFullModuleChildDir())
		deliveryHandlers := cliStageInputExistingDelivery(flagParam.getFullModuleChildDir())
		if flagParam.serviceName == "" {
			flagParam.serviceName = srvConfig.ServiceName
		}
		addUsecase(flagParam, ucName, deliveryHandlers)
		return

	case ApplyUsecase:
		if flagParam.isMonorepo {
			flagParam.serviceName = cliStageInputExistingServiceName(flagParam.outputFlag,
				"Please input existing service name:")
		}
		flagParam.moduleName = cliStageInputModule(filepath.Join(flagParam.outputFlag, flagParam.serviceName, "internal/modules"),
			"Please input existing module name:")
		ucName := cliStageInputExistingUsecaseName(flagParam.getFullModuleChildDir())
		deliveryHandlers := cliStageInputExistingDelivery(flagParam.getFullModuleChildDir())
		if flagParam.serviceName == "" {
			flagParam.serviceName = srvConfig.ServiceName
		}
		applyUsecaseToDelivery(flagParam, ucName, deliveryHandlers)
		return
	}

stageInputModules:
	cmdInput = readInput("Please input new module names (if more than one, separated by comma):")
	for _, moduleName := range strings.Split(cmdInput, ",") {
		path := "internal/modules/" + moduleName
		if flagParam.serviceName != "" {
			path = flagParam.outputFlag + flagParam.serviceName + "/" + path
		}
		if err := validateDir(path); scope != InitService && err == nil {
			fmt.Printf(RedFormat, "module '"+moduleName+"' is exist")
			goto stageInputModules
		}
		newModules = append(newModules, moduleConfig{
			ModuleName: strings.TrimSpace(candihelper.ToDelimited(moduleName, '-')), Skip: false,
		})
		flagParam.modules = append(flagParam.modules, strings.TrimSpace(moduleName))
	}
	if len(newModules) == 0 {
		fmt.Printf(RedFormat, "Modules cannot empty")
		goto stageInputModules
	}
	sort.Strings(flagParam.modules)
	srvConfig.Modules = append(srvConfig.Modules, newModules...)

stageSelectServerHandler:
	deliveryHandlerOption, deliveryHandlerMap = filterServerHandler(&srvConfig, flagParam)
	if len(deliveryHandlerMap) == 0 {
		goto stageSelectWorkerHandlers
	}
	cmdInput = readInput("Please select server handlers (separated by comma, enter for skip)\n" + deliveryHandlerOption)
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		opt := strings.TrimSpace(str)
		if serverName, ok := deliveryHandlerMap[opt]; ok {
			if handlers[serverName] {
				fmt.Printf(RedFormat, "Duplicate server handler type")
				goto stageSelectServerHandler
			}

			if serverName == RestHandler && scope == InitService {
			stageSelectRESTFramework:
				cmdInput = readInput("Please select REST library (choose one, enter for default using go-chi)",
					"1) Fiber REST API (plugin)")
				selected, ok := restPluginHandler[cmdInput]
				if ok {
					srvConfig.FiberRestHandler = selected == FiberRestDeps
				} else if cmdInput != "" {
					fmt.Printf(RedFormat, "Invalid option, try again")
					goto stageSelectRESTFramework
				}
			}

			handlers[serverName] = true
		} else if str != "" {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectServerHandler
		}
	}

stageSelectWorkerHandlers:
	deliveryHandlerOption, deliveryHandlerMap = filterWorkerHandler(&srvConfig, flagParam)
	cmdInput = readInput("Please select worker handlers (separated by comma, enter for skip)\n" + deliveryHandlerOption)
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if workerName, ok := deliveryHandlerMap[strings.TrimSpace(str)]; ok {
			handlers[workerName] = true
		} else if str != "" {
			fmt.Printf(RedFormat, "Invalid option, try again")
			goto stageSelectWorkerHandlers
		}
	}

	if len(handlers) == 0 {
		fmt.Printf(RedFormat, "No server/worker handler selected, try again")
		goto stageSelectServerHandler
	}

	srvConfig.workerPlugins = make(map[string]plugin)
	for k, v := range handlers {
		if !v {
			continue
		}
		switch k {
		case RestHandler:
			srvConfig.RestHandler = v
		case GrpcHandler:
			srvConfig.GRPCHandler = v
		case GraphqlHandler:
			srvConfig.GraphQLHandler = v
		case KafkaHandler:
			srvConfig.KafkaHandler = v
		case SchedulerHandler:
			srvConfig.SchedulerHandler = v
		case RedissubsHandler:
			srvConfig.RedisSubsHandler = v
		case TaskqueueHandler:
			srvConfig.TaskQueueHandler = v
		case PostgresListenerHandler:
			srvConfig.PostgresListenerHandler = v
		case RabbitmqHandler:
			srvConfig.RabbitMQHandler = v
		default:
			plg := plugins[k]
			if !slices.Contains(srvConfig.WorkerPlugins, k) {
				srvConfig.WorkerPlugins = append(srvConfig.WorkerPlugins, k)
			} else {
				plg.editAppFactory = nil
				plg.editConfig = nil
			}
			srvConfig.workerPlugins[k] = plg
		}
	}
	srvConfig.checkWorkerActive()

	for i := range srvConfig.Modules {
		srvConfig.Modules[i].config = srvConfig.config
		srvConfig.Modules[i].configHeader = srvConfig.configHeader
		srvConfig.Modules[i].RestHandler = handlers[RestHandler]
		srvConfig.Modules[i].GRPCHandler = handlers[GrpcHandler]
		srvConfig.Modules[i].GraphQLHandler = handlers[GraphqlHandler]
		srvConfig.Modules[i].KafkaHandler = handlers[KafkaHandler]
		srvConfig.Modules[i].SchedulerHandler = handlers[SchedulerHandler]
		srvConfig.Modules[i].RedisSubsHandler = handlers[RedissubsHandler]
		srvConfig.Modules[i].TaskQueueHandler = handlers[TaskqueueHandler]
		srvConfig.Modules[i].PostgresListenerHandler = handlers[PostgresListenerHandler]
		srvConfig.Modules[i].RabbitMQHandler = handlers[RabbitmqHandler]
		srvConfig.Modules[i].WorkerPlugins = srvConfig.WorkerPlugins
		srvConfig.Modules[i].checkWorkerActive()
	}

	if scope == AddModule || scope == AddHandler {
		srvConfig.OutputDir = ""
		if scope == AddHandler {
			srvConfig.parseDefaultHeader()
			scopeAddHandler(flagParam, srvConfig, handlers, handlers)
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
	if handlers[RedissubsHandler] && !dependencies[RedisDeps] {
		fmt.Printf(RedFormat, "Redis Subscriber need redis, try again")
		goto stageSelectDependencies
	}

	if dependencies[SqldbDeps] {
	stageSelectSQLDriver:
		cmdInput = readInput("Please select SQL database driver (choose one)\n" +
			"1) Postgres\n" +
			"2) MySQL\n" +
			"3) SQLite3")
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

	if handlers[PostgresListenerHandler] && (!dependencies[SqldbDeps] || srvConfig.SQLDriver != "postgres") {
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
	srvConfig.RedisDeps = dependencies[RedisDeps]
	srvConfig.SQLDeps, srvConfig.MongoDeps, srvConfig.ArangoDeps = dependencies[SqldbDeps], dependencies[MongodbDeps], dependencies[ArangodbDeps]
	srvConfig.checkWorkerActive()

	return
}
