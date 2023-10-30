package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func scopeAddHandler(flagParam *flagParameter, cfg serviceConfig, serverHandlers, workerHandler map[string]bool) {
	var mod moduleConfig
	gqlCfg := cfg
	gqlCfg.Modules = nil
	for _, m := range cfg.Modules {
		if flagParam.moduleName == m.ModuleName {
			mod = m
			gqlCfg.Modules = append(gqlCfg.Modules, m)
			flagParam.modules = append(flagParam.modules, m.ModuleName)
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
			{FromTemplate: true, DataSource: gqlCfg, Source: defaultGraphqlRootSchema, FileName: "_schema.graphql", SkipIfExist: true},
			{FromTemplate: true, DataSource: gqlCfg, SkipIfExist: true, Source: templateGraphqlCommon, FileName: "_common.graphql"},
		},
	}
	deliveryStructure := FileStructure{
		TargetDir: "delivery/", IsDir: true, SkipIfExist: true,
	}

	replaceEnv, replaceConfigs, replaceMainModule := make(map[string]string), make(map[string]string), make(map[string]string)
	deliveryPackageDir := fmt.Sprintf(`"%s/internal/modules/%s/delivery`, mod.PackagePrefix, mod.ModuleName)
	for handler := range serverHandlers {
		switch handler {
		case RestHandler:
			deliveryStructure.Childs = append(deliveryStructure.Childs, FileStructure{
				TargetDir: "resthandler/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: deliveryRestTemplate, FileName: "resthandler.go"},
				},
			})
			replaceEnv["USE_REST=false"] = "USE_REST=true"
			replaceMainModule["// mod.restHandler"] = "mod.restHandler"
			replaceMainModule["// "+deliveryPackageDir+"/resthandler"] = deliveryPackageDir + "/resthandler"
		case GrpcHandler:
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
			replaceMainModule["// mod.grpcHandler"] = "mod.grpcHandler"
			replaceMainModule["// "+deliveryPackageDir+"/grpchandler"] = deliveryPackageDir + "/grpchandler"
		case GraphqlHandler:
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
				},
			})
			replaceEnv["USE_GRAPHQL=false"] = "USE_GRAPHQL=true"
			replaceMainModule["// mod.graphqlHandler"] = "mod.graphqlHandler"
			replaceMainModule["// "+deliveryPackageDir+"/graphqlhandler"] = deliveryPackageDir + "/graphqlhandler"
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
	mod.PostgresListenerHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/postgres_listener_handler.go", "/"))
	mod.RabbitMQHandler = isDirExist(strings.TrimPrefix(flagParam.outputFlag+flagParam.serviceName+"/internal/modules/"+mod.ModuleName+"/delivery/workerhandler/rabbitmq_handler.go", "/"))
	for handler := range workerHandler {
		switch handler {
		case KafkaHandler:
			mod.KafkaHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryKafkaTemplate, FileName: "kafka_handler.go",
			})
			replaceEnv["USE_KAFKA_CONSUMER=false"] = "USE_KAFKA_CONSUMER=true"
			replaceConfigs["// broker.SetKafka(broker.NewKafkaBroker())"] = "	broker.SetKafka(broker.NewKafkaBroker())"
			replaceMainModule["// types.Kafka"] = "types.Kafka"
			replaceMainModule["// "+deliveryPackageDir+"/workerhandler"] = deliveryPackageDir + "/workerhandler"
		case SchedulerHandler:
			mod.SchedulerHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryCronTemplate, FileName: "cron_handler.go",
			})
			replaceEnv["USE_CRON_SCHEDULER=false"] = "USE_CRON_SCHEDULER=true"
			replaceMainModule["// types.Scheduler"] = "types.Scheduler"
			replaceMainModule["// "+deliveryPackageDir+"/workerhandler"] = deliveryPackageDir + "/workerhandler"
		case RedissubsHandler:
			mod.RedisSubsHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryRedisTemplate, FileName: "redis_handler.go",
			})
			replaceEnv["USE_REDIS_SUBSCRIBER=false"] = "USE_REDIS_SUBSCRIBER=true"
			replaceMainModule["// types.RedisSubscriber"] = "types.RedisSubscriber"
		case TaskqueueHandler:
			mod.TaskQueueHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryTaskQueueTemplate, FileName: "taskqueue_handler.go",
			})
			replaceEnv["USE_TASK_QUEUE_WORKER=false"] = "USE_TASK_QUEUE_WORKER=true"
			replaceMainModule["// types.TaskQueue"] = "types.TaskQueue"
			replaceMainModule["// "+deliveryPackageDir+"/workerhandler"] = deliveryPackageDir + "/workerhandler"
		case PostgresListenerHandler:
			mod.PostgresListenerHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryPostgresListenerTemplate, FileName: "postgres_listener_handler.go",
			})
			replaceEnv["USE_POSTGRES_LISTENER_WORKER=false"] = "USE_POSTGRES_LISTENER_WORKER=true"
			replaceMainModule["// types.PostgresListener"] = "types.PostgresListener"
			replaceMainModule["// "+deliveryPackageDir+"/workerhandler"] = deliveryPackageDir + "/workerhandler"
		case RabbitmqHandler:
			mod.RabbitMQHandler = true
			deliveryWorkerStructure.Childs = append(deliveryWorkerStructure.Childs, FileStructure{
				FromTemplate: true, DataSource: mod, Source: deliveryRabbitMQTemplate, FileName: "rabbitmq_handler.go",
			})
			replaceEnv["USE_RABBITMQ_CONSUMER=false"] = "USE_RABBITMQ_CONSUMER=true"
			replaceConfigs["// broker.SetRabbitMQ(broker.NewRabbitMQBroker())"] = "	broker.SetRabbitMQ(broker.NewRabbitMQBroker())"
			replaceMainModule["// types.RabbitMQ"] = "types.RabbitMQ"
			replaceMainModule["// "+deliveryPackageDir+"/workerhandler"] = deliveryPackageDir + "/workerhandler"
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
		},
	}
	internalServiceStructure := FileStructure{
		TargetDir: "internal/", IsDir: true, Skip: true, Childs: []FileStructure{
			{TargetDir: "modules/", IsDir: true, Skip: true, Childs: []FileStructure{moduleStructure}},
		},
	}

	root := FileStructure{
		Skip: true, Childs: []FileStructure{
			apiStructure, internalServiceStructure, {
				Source: gqlCfg.toJSONString(), FileName: "candi.json",
			},
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
	for old, new := range replaceMainModule {
		readFileAndApply(root.TargetDir+"internal/modules/"+mod.ModuleName+"/module.go", old, new)
	}
	if serverHandlers[GraphqlHandler] {
		updateGraphQLRoot(flagParam, gqlCfg)
	}
}

func generateServiceSDK(srvConfig serviceConfig) {
	b, err := os.ReadFile("sdk/sdk.go")
	if err != nil {
		return
	}

	b = bytes.Replace(b, []byte("@candi:serviceImport"), []byte(fmt.Sprintf("@candi:serviceImport\n	\"monorepo/sdk/%s\"", srvConfig.ServiceName)), -1)
	b = bytes.Replace(b, []byte("@candi:construct"), loadTemplate("@candi:construct\n\n"+`// Set{{upper (clean $.ServiceName)}} option func
func Set{{upper (clean $.ServiceName)}}({{lower (clean $.ServiceName)}} {{lower (clean $.ServiceName)}}.{{upper (clean $.ServiceName)}}) Option {
	return func(s *sdkInstance) {
		s.{{lower (clean $.ServiceName)}} = {{lower (clean $.ServiceName)}}
	}
}`, srvConfig), -1)
	b = bytes.Replace(b, []byte("@candi:serviceMethod"), loadTemplate("@candi:serviceMethod\n	"+`{{upper (clean $.ServiceName)}}() {{lower (clean $.ServiceName)}}.{{upper (clean $.ServiceName)}}`, srvConfig), -1)
	b = bytes.Replace(b, []byte("@candi:serviceField"), loadTemplate("@candi:serviceField\n	{{lower (clean $.ServiceName)}}	{{lower (clean $.ServiceName)}}.{{upper (clean $.ServiceName)}}", srvConfig), -1)
	b = bytes.Replace(b, []byte("@candi:instanceMethod"), loadTemplate("@candi:instanceMethod\n"+`func (s *sdkInstance) {{upper (clean $.ServiceName)}}() {{lower (clean $.ServiceName)}}.{{upper (clean $.ServiceName)}} {
	return s.{{lower (clean $.ServiceName)}}
}`, srvConfig), -1)
	os.WriteFile("sdk/sdk.go", b, 0644)

	var fileStructure FileStructure
	fileStructure.Skip = true
	fileStructure.TargetDir = "sdk/"
	fileStructure.Childs = []FileStructure{
		{TargetDir: srvConfig.ServiceName + "/", IsDir: true, SkipIfExist: true, Childs: []FileStructure{
			{FromTemplate: true, DataSource: srvConfig, Source: templateSDKServiceAbstraction, FileName: srvConfig.ServiceName + ".go"},
			{FromTemplate: true, DataSource: srvConfig, Source: templateSDKServiceGRPC, FileName: srvConfig.ServiceName + "_grpc.go"},
			{FromTemplate: true, DataSource: srvConfig, Source: templateSDKServiceREST, FileName: srvConfig.ServiceName + "_rest.go"},
		}},
	}
	execGenerator(fileStructure)

	// generate global shared
	b, err = os.ReadFile("globalshared/gorm_tracer.go")
	if err != nil && srvConfig.SQLUseGORM {
		globalShared := FileStructure{
			TargetDir: "globalshared/", IsDir: true, SkipIfExist: true, Childs: []FileStructure{
				{FromTemplate: true, DataSource: srvConfig,
					Source: templateGORMTracer, FileName: "gorm_tracer.go"},
			},
		}
		execGenerator(globalShared)
	}
}

func updateGraphQLRoot(flagParam *flagParameter, cfg serviceConfig) {
	path := "api/graphql/_schema.graphql"
	if flagParam.serviceName != "" {
		path = flagParam.outputFlag + flagParam.serviceName + "/" + path
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, moduleName := range flagParam.modules {
		cleanMod, cleanUpperMod := candihelper.ToCamelCase(moduleName), strings.Title(candihelper.ToCamelCase(moduleName))
		b = bytes.ReplaceAll(b, []byte("@candi:queryRoot"), []byte(fmt.Sprintf("@candi:queryRoot\n	%s: %sQueryResolver @auth(authType: BEARER)", cleanMod, cleanUpperMod)))
		b = bytes.ReplaceAll(b, []byte("@candi:mutationRoot"), []byte(fmt.Sprintf("@candi:mutationRoot\n	%s: %sMutationResolver @auth(authType: BEARER)", cleanMod, cleanUpperMod)))
		b = bytes.ReplaceAll(b, []byte("@candi:subscriptionRoot"), []byte(fmt.Sprintf("@candi:subscriptionRoot\n	%s: %sSubscriptionResolver", cleanMod, cleanUpperMod)))
	}
	os.WriteFile(path, b, 0644)
}

func updateSharedUsecase(flagParam flagParameter, cfg serviceConfig) {
	path := "pkg/shared/usecase/usecase.go"
	if flagParam.serviceName != "" {
		path = flagParam.outputFlag + flagParam.serviceName + "/" + path
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, moduleName := range flagParam.modules {
		cleanMod, cleanPathMod := strings.ToLower(candihelper.ToCamelCase(moduleName)), strings.ToLower(candihelper.ToDelimited(moduleName, '-'))
		cleanUpperMod := strings.Title(candihelper.ToCamelCase(moduleName))
		b = bytes.Replace(b, []byte("@candi:usecaseImport"),
			[]byte(fmt.Sprintf("@candi:usecaseImport\n	%susecase \"%s/internal/modules/%s/usecase\"", cleanMod, cfg.PackagePrefix, cleanPathMod)), -1)
		b = bytes.Replace(b, []byte("@candi:usecaseMethod"),
			[]byte(fmt.Sprintf("@candi:usecaseMethod\n		%s() %susecase.%sUsecase", cleanUpperMod, cleanMod, cleanUpperMod)), -1)
		b = bytes.Replace(b, []byte("@candi:usecaseField"),
			[]byte(fmt.Sprintf("@candi:usecaseField\n		%susecase.%sUsecase", cleanMod, cleanUpperMod)), -1)
		b = bytes.Replace(b, []byte("@candi:usecaseCommon"),
			[]byte(fmt.Sprintf("@candi:usecaseCommon\n		usecaseInst.%sUsecase, setSharedUsecaseFunc = %susecase.New%sUsecase(deps)\n"+
				"		setSharedUsecaseFuncs = append(setSharedUsecaseFuncs, setSharedUsecaseFunc)", cleanUpperMod, cleanMod, cleanUpperMod)), -1)
		b = bytes.Replace(b, []byte("@candi:usecaseImplementation"),
			[]byte(fmt.Sprintf("@candi:usecaseImplementation\n"+`func (uc *usecaseUow) %s() %susecase.%sUsecase {
	return uc.%sUsecase
}
`, cleanUpperMod, cleanMod, cleanUpperMod, cleanUpperMod)), -1)
	}

	os.WriteFile(path, b, 0644)
}

func updateSharedRepository(flagParam flagParameter, cfg serviceConfig) {
	for repoType, repo := range map[string]string{"SQL": "repository_sql.go", "Mongo": "repository_mongo.go", "Arango": "repository_arango.go"} {
		path := "pkg/shared/repository/" + repo
		if flagParam.serviceName != "" {
			path = flagParam.outputFlag + flagParam.serviceName + "/" + path
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, moduleName := range flagParam.modules {
			cleanMod, cleanPathMod := strings.ToLower(candihelper.ToCamelCase(moduleName)), strings.ToLower(candihelper.ToDelimited(moduleName, '-'))
			cleanUpperMod := strings.Title(candihelper.ToCamelCase(moduleName))
			b = bytes.Replace(b, []byte("@candi:repositoryImport"),
				[]byte(fmt.Sprintf("@candi:repositoryImport\n	%srepo \"%s/internal/modules/%s/repository\"", cleanMod, cfg.PackagePrefix, cleanPathMod)), -1)
			b = bytes.Replace(b, []byte("@candi:repositoryMethod"),
				[]byte(fmt.Sprintf("@candi:repositoryMethod\n		%sRepo() %srepo.%sRepository", cleanUpperMod, cleanMod, cleanUpperMod)), -1)
			b = bytes.Replace(b, []byte("@candi:repositoryField"),
				[]byte(fmt.Sprintf("@candi:repositoryField\n		%sRepo %srepo.%sRepository", cleanMod, cleanMod, cleanUpperMod)), -1)

			if repoType == "SQL" && cfg.SQLDeps {
				b = bytes.Replace(b, []byte("@candi:repositoryConstructor"),
					[]byte(fmt.Sprintf("@candi:repositoryConstructor\n		%sRepo: %srepo.New%sRepoSQL(readDB, writeDB),", cleanMod, cleanMod, cleanUpperMod)), -1)
				b = bytes.Replace(b, []byte("@candi:repositoryDestructor"),
					[]byte(fmt.Sprintf("@candi:repositoryDestructor\n	r.%sRepo = nil", cleanMod)), -1)
			} else if repoType == "Mongo" && cfg.MongoDeps {
				b = bytes.Replace(b, []byte("@candi:repositoryConstructor"),
					[]byte(fmt.Sprintf("@candi:repositoryConstructor\n		%sRepo: %srepo.New%sRepoMongo(readDB, writeDB),", cleanMod, cleanMod, cleanUpperMod)), -1)
			} else if repoType == "Arango" && cfg.ArangoDeps {
				b = bytes.Replace(b, []byte("@candi:repositoryConstructor"),
					[]byte(fmt.Sprintf("@candi:repositoryConstructor\n		%sRepo: %srepo.New%sRepoArango(readDB, writeDB),", cleanMod, cleanMod, cleanUpperMod)), -1)
			}
			b = bytes.Replace(b, []byte("@candi:repositoryImplementation"),
				[]byte(fmt.Sprintf("@candi:repositoryImplementation\n"+`func (r *repo%sImpl) %sRepo() %srepo.%sRepository {
	return r.%sRepo
}
`, repoType, cleanUpperMod, cleanMod, cleanUpperMod, cleanMod)), -1)
		}

		os.WriteFile(path, b, 0644)
	}
}
