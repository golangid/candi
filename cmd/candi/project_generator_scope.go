package main

import (
	"encoding/json"
	"log"
	"strings"
)

func scopeAddHandler(flagParam *flagParameter, cfg serviceConfig, serverHandlers, workerHandler map[string]bool) {
	var mod moduleConfig
	gqlCfg := cfg
	gqlCfg.Modules = nil
	for _, m := range cfg.Modules {
		if flagParam.moduleName == m.ModuleName {
			mod = m
			gqlCfg.Modules = append(gqlCfg.Modules, m)
			continue
		}

		if isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+m.ModuleName+"/delivery/graphqlhandler", "/")) {
			m.configHeader = cfg.configHeader
			m.config = cfg.config
			gqlCfg.Modules = append(gqlCfg.Modules, m)
		}
	}

	if mod.ModuleName == "" {
		log.Fatal("module is empty")
	}

	mod.configHeader = cfg.configHeader
	mod.config = cfg.config

	apiProtoStructure := FileStructure{
		TargetDir: "proto/", IsDir: true, SkipIfExist: true,
	}
	apiGraphQLStructure := FileStructure{
		TargetDir: "graphql/", IsDir: true, SkipIfExist: true, Childs: []FileStructure{
			{FromTemplate: true, DataSource: gqlCfg, Source: defaultGraphqlRootSchema, FileName: "_schema.graphql"},
			{FromTemplate: true, DataSource: gqlCfg, SkipIfExist: true, Source: templateGraphqlCommon, FileName: "_common.graphql"},
		},
	}
	deliveryStructure := FileStructure{
		TargetDir: "delivery/", IsDir: true, SkipIfExist: true,
	}

	replaceEnv, replaceConfigs := make(map[string]string), make(map[string]string)
	for handler := range serverHandlers {
		switch handler {
		case restHandler:
			deliveryStructure.Childs = append(deliveryStructure.Childs, FileStructure{
				TargetDir: "resthandler/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: deliveryRestTemplate, FileName: "resthandler.go"},
				},
			})
			replaceEnv["USE_REST=false"] = "USE_REST=true"
		case grpcHandler:
			apiProtoStructure.Childs = append(apiProtoStructure.Childs, FileStructure{
				TargetDir: mod.ModuleName + "/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: defaultGRPCProto, FileName: mod.ModuleName + ".proto"},
				},
			})
			deliveryStructure.Childs = append(deliveryStructure.Childs, FileStructure{
				TargetDir: "grpchandler/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: deliveryGRPCTemplate, FileName: "grpchandler.go"},
				},
			})
			replaceEnv["USE_GRPC=false"] = "USE_GRPC=true"
		case graphqlHandler:
			apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: defaultGraphqlSchema, FileName: mod.ModuleName + ".graphql",
			})
			deliveryStructure.Childs = append(deliveryStructure.Childs, FileStructure{
				TargetDir: "graphqlhandler/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlRootTemplate, FileName: "root_resolver.go"},
					{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlQueryTemplate, FileName: "query_resolver.go"},
					{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlMutationTemplate, FileName: "mutation_resolver.go"},
					{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlSubscriptionTemplate, FileName: "subscription_resolver.go"},
					{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlFieldResolverTemplate, FileName: "field_serializer_resolver.go"},
				},
			})
			replaceEnv["USE_GRAPHQL=false"] = "USE_GRAPHQL=true"
		}
	}

	if len(apiProtoStructure.Childs) == 0 {
		apiProtoStructure.Skip = true
	}
	if len(apiGraphQLStructure.Childs) == 2 {
		apiGraphQLStructure.Childs[0].Skip = true
		apiGraphQLStructure.Childs[1].Skip = true
	}

	deliveryWorkerStructure := FileStructure{
		TargetDir: "workerhandler/", IsDir: true, SkipIfExist: true, Skip: !cfg.IsWorkerActive,
	}
	mod.KafkaHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/kafka_handler.go", "/"))
	mod.SchedulerHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/cron_handler.go", "/"))
	mod.RedisSubsHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/redis_handler.go", "/"))
	mod.TaskQueueHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/taskqueue_handler.go", "/"))
	for handler := range workerHandler {
		switch handler {
		case kafkaHandler:
			mod.KafkaHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryKafkaTemplate, FileName: "kafka_handler.go",
			})
			replaceEnv["USE_KAFKA_CONSUMER=false"] = "USE_KAFKA_CONSUMER=true"
			replaceConfigs["// \""+mod.LibraryName+"/codebase/factory/types\""] = "\"" + mod.LibraryName + "/codebase/factory/types\""
			replaceConfigs["// types.Kafka"] = "types.Kafka"
		case schedulerHandler:
			mod.SchedulerHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryCronTemplate, FileName: "cron_handler.go",
			})
			replaceEnv["USE_CRON_SCHEDULER=false"] = "USE_CRON_SCHEDULER=true"
		case redissubsHandler:
			mod.RedisSubsHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryRedisTemplate, FileName: "redis_handler.go",
			})
			replaceEnv["USE_REDIS_SUBSCRIBER=false"] = "USE_REDIS_SUBSCRIBER=true"
		case taskqueueHandler:
			mod.TaskQueueHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryTaskQueueTemplate, FileName: "taskqueue_handler.go",
			})
			replaceEnv["USE_TASK_QUEUE_WORKER=false"] = "USE_TASK_QUEUE_WORKER=true"
		case postgresListenerHandler:
			mod.PostgresListenerHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryPostgresListenerTemplate, FileName: "postgres_listener_handler.go",
			})
			replaceEnv["USE_POSTGRES_LISTENER_WORKER=false"] = "USE_POSTGRES_LISTENER_WORKER=true"
		}
	}
	apiStructure := FileStructure{
		TargetDir: "api/", IsDir: true, Skip: true, Childs: []FileStructure{
			apiProtoStructure, apiGraphQLStructure,
		},
	}
	deliveryStructure.Childs = append(deliveryStructure.Childs, deliveryWorkerStructure)
	moduleStructure := FileStructure{
		TargetDir: mod.ModuleName + "/", IsDir: true, Skip: true,
		Childs: []FileStructure{
			deliveryStructure,
			{FromTemplate: true, DataSource: mod, Source: moduleMainTemplate, FileName: "module.go"},
		},
	}
	internalServiceStructure := FileStructure{
		TargetDir: "internal/", IsDir: true, Skip: true, Childs: []FileStructure{
			{TargetDir: "modules/", IsDir: true, Skip: true, Childs: []FileStructure{moduleStructure}},
		},
	}

	configJSON, _ := json.Marshal(cfg)
	root := FileStructure{
		Skip: true, Childs: []FileStructure{
			apiStructure, internalServiceStructure,
			{Source: string(configJSON), FileName: "candi.json"},
		},
	}

	if isWorkdirMonorepo() {
		root.TargetDir = cfg.OutputDir + cfg.ServiceName + "/"
	}

	execGenerator(root)

	for old, new := range replaceEnv {
		readFileAndApply(root.TargetDir+".env", old, new)
		readFileAndApply(root.TargetDir+".env.sample", old, new)
	}
	for old, new := range replaceConfigs {
		readFileAndApply(root.TargetDir+"configs/configs.go", old, new)
	}
}
