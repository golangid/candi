package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/golangid/candi"
	"github.com/golangid/candi/candihelper"
)

func projectGenerator(flagParam flagParameter, scope string, srvConfig serviceConfig) {
	if flagParam.addHandler {
		return
	}

	sort.Slice(srvConfig.Modules, func(i, j int) bool {
		return srvConfig.Modules[i].ModuleName < srvConfig.Modules[j].ModuleName
	})

	apiStructure := FileStructure{
		TargetDir: "api/", IsDir: true,
	}
	internalServiceStructure := FileStructure{
		TargetDir: "internal/", IsDir: true,
	}
	pkgServiceStructure := FileStructure{
		TargetDir: "pkg/", IsDir: true, DataSource: srvConfig,
	}
	apiProtoStructure := FileStructure{
		TargetDir: "proto/", IsDir: true,
	}
	apiGraphQLStructure := FileStructure{
		TargetDir: "graphql/", IsDir: true, SkipFunc: func() bool { return !srvConfig.GraphQLHandler },
	}
	apiJSONSchemaStructure := FileStructure{
		TargetDir: "jsonschema/", IsDir: true,
	}

	apiGraphQLSchemaStructure := FileStructure{
		FromTemplate: true, DataSource: srvConfig, Source: defaultGraphqlRootSchema, FileName: "_schema.graphql",
		SkipFunc: func() bool { return !srvConfig.GraphQLHandler },
	}

	moduleStructure := FileStructure{
		TargetDir: "modules/", IsDir: true, DataSource: srvConfig,
	}
	var sharedDomainFiles, migrationFiles []FileStructure

	for _, mod := range srvConfig.Modules {
		mod.configHeader = srvConfig.configHeader
		mod.config = srvConfig.config
		if mod.Skip {
			continue
		}

		var repoModule = []FileStructure{
			{FromTemplate: true, DataSource: mod, Source: templateRepositoryAbstraction, FileName: "repository.go"},
		}
		repoModule = append(repoModule, parseRepositoryModule(mod)...)

		cleanArchModuleDir := []FileStructure{
			{
				TargetDir: "delivery/", IsDir: true,
				Childs: []FileStructure{
					{TargetDir: "graphqlhandler/", IsDir: true, SkipFunc: func() bool { return !srvConfig.GraphQLHandler },
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlRootTemplate, FileName: "root_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlQueryTemplate, FileName: "query_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlMutationTemplate, FileName: "mutation_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlSubscriptionTemplate, FileName: "subscription_resolver.go"},
						}},
					{TargetDir: "grpchandler/", IsDir: true, SkipFunc: func() bool { return !srvConfig.GRPCHandler },
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryGRPCTemplate, FileName: "grpchandler.go"},
						}},
					{TargetDir: "resthandler/", IsDir: true, SkipFunc: func() bool { return !srvConfig.RestHandler },
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryRestTemplate, FileName: "resthandler.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryRestTestTemplate, FileName: "resthandler_test.go"},
						}},
					{TargetDir: "workerhandler/", IsDir: true, Skip: !srvConfig.IsWorkerActive,
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryKafkaTemplate, FileName: "kafka_handler.go",
								SkipFunc: func() bool { return !srvConfig.KafkaHandler }},
							{FromTemplate: true, DataSource: mod, Source: deliveryRedisTemplate, FileName: "redis_handler.go",
								SkipFunc: func() bool { return !srvConfig.RedisSubsHandler }},
							{FromTemplate: true, DataSource: mod, Source: deliveryCronTemplate, FileName: "cron_handler.go",
								SkipFunc: func() bool { return !srvConfig.SchedulerHandler }},
							{FromTemplate: true, DataSource: mod, Source: deliveryTaskQueueTemplate, FileName: "taskqueue_handler.go",
								SkipFunc: func() bool { return !srvConfig.TaskQueueHandler }},
							{FromTemplate: true, DataSource: mod, Source: deliveryPostgresListenerTemplate, FileName: "postgres_listener_handler.go",
								SkipFunc: func() bool { return !srvConfig.PostgresListenerHandler }},
							{FromTemplate: true, DataSource: mod, Source: deliveryRabbitMQTemplate, FileName: "rabbitmq_handler.go",
								SkipFunc: func() bool { return !srvConfig.RabbitMQHandler }},
						}},
				},
			},
			{
				TargetDir: "domain/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: templateModuleDomain, FileName: "filter.go"},
					{FromTemplate: true, DataSource: mod, Source: "package domain\n", FileName: "payload.go"},
					{FromTemplate: true, DataSource: mod, Source: templateModuleRequestDomain, FileName: "request.go"},
					{FromTemplate: true, DataSource: mod, Source: templateModuleResponseDomain, FileName: "response.go"},
				},
			},
			{
				TargetDir: "repository/", IsDir: true,
				Childs: repoModule,
			},
			{
				TargetDir: "usecase/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseAbstraction, FileName: "usecase.go"},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseTest, FileName: "usecase_test.go"},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseGetAll, FileName: fmt.Sprintf("get_all_%s.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseGetAllTest, FileName: fmt.Sprintf("get_all_%s_test.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseGetDetail, FileName: fmt.Sprintf("get_detail_%s.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseGetDetailTest, FileName: fmt.Sprintf("get_detail_%s_test.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseCreate, FileName: fmt.Sprintf("create_%s.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseCreateTest, FileName: fmt.Sprintf("create_%s_test.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseUpdate, FileName: fmt.Sprintf("update_%s.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseUpdateTest, FileName: fmt.Sprintf("update_%s_test.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseDelete, FileName: fmt.Sprintf("delete_%s.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseDeleteTest, FileName: fmt.Sprintf("delete_%s_test.go", candihelper.ToDelimited(mod.ModuleName, '_'))},
				},
			},
		}

		moduleStructure.Childs = append(moduleStructure.Childs, []FileStructure{
			{
				TargetDir: mod.ModuleName + "/", IsDir: true,
				Childs: append(cleanArchModuleDir,
					FileStructure{
						FromTemplate: true, DataSource: mod, Source: moduleMainTemplate, FileName: "module.go",
					},
				),
			},
		}...)

		apiProtoStructure.Childs = append(apiProtoStructure.Childs, FileStructure{
			TargetDir: mod.ModuleName + "/", IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, DataSource: mod, Source: defaultGRPCProto, FileName: mod.ModuleName + ".proto"},
			},
			SkipFunc: func() bool { return !srvConfig.GRPCHandler },
		})
		apiJSONSchemaStructure.Childs = append(apiJSONSchemaStructure.Childs, FileStructure{
			TargetDir: mod.ModuleName + "/", IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, DataSource: mod, Source: jsonSchemaFilterGetTemplate, FileName: "get_all.json"},
				{FromTemplate: true, DataSource: mod, Source: jsonSchemaSaveTemplate, FileName: "save.json"},
			},
		})
		apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, FileStructure{
			FromTemplate: true, DataSource: mod, Source: defaultGraphqlSchema, FileName: mod.ModuleName + ".graphql",
			SkipFunc: func() bool { return !srvConfig.GraphQLHandler },
		})

		// for shared domain
		sharedDomainFiles = append(sharedDomainFiles, FileStructure{
			FromTemplate: true, DataSource: mod, Source: templateSharedDomain, FileName: candihelper.ToDelimited(mod.ModuleName, '_') + ".go",
		})
		migrationFiles = append(migrationFiles, FileStructure{
			FromTemplate: true, DataSource: mod, Source: templateCmdMigrationInitModule,
			FileName: time.Now().Format("20060102150405") + "_create_table_" +
				candihelper.ToDelimited(pluralize.NewClient().Plural(mod.ModuleName), '_') + ".sql",
			SkipFunc: func() bool {
				return !srvConfig.SQLDeps
			},
		})
	}

	configsStructure := FileStructure{
		TargetDir: "configs/", IsDir: true,
		Childs: []FileStructure{
			{FromTemplate: true, DataSource: srvConfig, Source: appFactoryTemplate, FileName: "app_factory.go"},
			{FromTemplate: true, DataSource: srvConfig, Source: configsTemplate, FileName: "configs.go"},
		},
	}

	internalServiceStructure.Childs = append(internalServiceStructure.Childs, FileStructure{
		FromTemplate: true, DataSource: srvConfig, Source: serviceMainTemplate, FileName: "service.go",
	})

	var baseDirectoryFile FileStructure
	baseDirectoryFile.TargetDir = flagParam.outputFlag + srvConfig.ServiceName + "/"
	baseDirectoryFile.DataSource = srvConfig
	baseDirectoryFile.IsDir = true
	switch scope {
	case InitService:
		apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, []FileStructure{
			apiGraphQLSchemaStructure,
			{FromTemplate: true, DataSource: srvConfig, Source: templateGraphqlCommon, FileName: "_common.graphql",
				SkipFunc: func() bool { return !srvConfig.GraphQLHandler }},
		}...)
		apiStructure.Childs = []FileStructure{
			apiGraphQLStructure,
			apiProtoStructure,
			apiJSONSchemaStructure,
		}
		migrationFiles = append(migrationFiles, FileStructure{FromTemplate: true, DataSource: srvConfig,
			Source: templateCmdMigrationInitTable, FileName: "00000000000000_init_tables.go"})
		cmdStructure := FileStructure{
			TargetDir: "cmd/", IsDir: true,
			Childs: []FileStructure{
				{TargetDir: srvConfig.ServiceName + "/", IsDir: true, DataSource: srvConfig},
				{TargetDir: "migration/", IsDir: true, Childs: []FileStructure{
					{FileName: ".gitkeep"},
					{FromTemplate: true, DataSource: srvConfig, Source: templateCmdMigration, FileName: "migration.go",
						SkipFunc: func() bool {
							return !srvConfig.SQLDeps
						}},
					{TargetDir: "migrations/", IsDir: true, Childs: migrationFiles,
						SkipFunc: func() bool {
							return !srvConfig.SQLDeps
						}},
				}},
			},
		}
		internalServiceStructure.Childs = append(internalServiceStructure.Childs, moduleStructure)
		pkgServiceStructure.Childs = append(pkgServiceStructure.Childs, []FileStructure{
			{TargetDir: "helper/", IsDir: true, Childs: []FileStructure{
				{FromTemplate: true, FileName: "helper.go"},
			}},
			{TargetDir: "shared/", IsDir: true, Childs: []FileStructure{
				{TargetDir: "domain/", IsDir: true, Childs: sharedDomainFiles},
				{TargetDir: "repository/", IsDir: true, Childs: parseSharedRepository(srvConfig)},
				{TargetDir: "usecase/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, DataSource: srvConfig, Source: templateUsecaseUOW, FileName: "usecase.go"},
					{TargetDir: "common/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, DataSource: srvConfig, Source: templateUsecaseCommon, FileName: "common.go"},
					}},
				}},
				{FromTemplate: true, DataSource: srvConfig, Source: templateSharedMiddlewareImpl, FileName: "middleware_impl_example.go"},
				{Source: additionalEnvTemplate, FromTemplate: true, DataSource: srvConfig, FileName: "environment.go"},
				{Source: templateGORMTracer, Skip: !(!flagParam.isMonorepo && srvConfig.SQLUseGORM), FromTemplate: true,
					DataSource: srvConfig, FileName: "gorm_tracer.go"},
			}},
		}...)
		baseDirectoryFile.Childs = []FileStructure{
			apiStructure, cmdStructure, configsStructure, internalServiceStructure, pkgServiceStructure,
			{TargetDir: "deployments/", IsDir: true, Childs: []FileStructure{
				{TargetDir: "k8s/", IsDir: true, Childs: []FileStructure{
					{FileName: srvConfig.ServiceName + ".yaml"},
				}},
			}},
			{TargetDir: "docs/", IsDir: true, Childs: []FileStructure{
				{FileName: ".gitkeep"},
			}},
			{FromTemplate: true, DataSource: srvConfig, Source: gitignoreTemplate, FileName: ".gitignore"},
			{FromTemplate: true, DataSource: srvConfig, Source: makefileTemplate, FileName: "Makefile"},
			{FromTemplate: true, DataSource: srvConfig, Source: dockerfileTemplate, FileName: "Dockerfile", SkipFunc: func() bool {
				return isWorkdirMonorepo()
			}},
			{FromTemplate: true, DataSource: srvConfig, Source: cmdMainTemplate, FileName: "main.go"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env.sample"},
			{Source: srvConfig.toJSONString(), FileName: "candi.json"},
			{FromTemplate: true, DataSource: srvConfig, Source: readmeTemplate, FileName: "README.md"},
			{FromTemplate: true, DataSource: srvConfig, Source: licenseMapTemplate[srvConfig.License], FileName: "LICENSE",
				Skip: srvConfig.License == ""},
		}
		if flagParam.withGoModFlag {
			baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, FileStructure{
				FromTemplate: true, DataSource: srvConfig, Source: gomodTemplate, FileName: "go.mod",
			})
		}

		if flagParam.isMonorepo {
			generateServiceSDK(srvConfig)
		}

	case AddModule:
		cmdStructure := FileStructure{
			TargetDir: "cmd/", IsDir: true, Skip: true, Childs: []FileStructure{
				{TargetDir: "migration/", IsDir: true, Skip: true, Childs: []FileStructure{
					{TargetDir: "migrations/", IsDir: true, Skip: true, Childs: migrationFiles},
				}, SkipFunc: func() bool {
					return !srvConfig.SQLDeps
				}},
			},
		}
		configsStructure.Skip = true
		moduleStructure.Skip = true
		pkgServiceStructure.Skip = true
		internalServiceStructure.Skip = true
		pkgServiceStructure.Childs = []FileStructure{
			{TargetDir: "shared/", IsDir: true, Skip: true, Childs: []FileStructure{
				{TargetDir: "domain/", IsDir: true, Skip: true, Childs: sharedDomainFiles},
			}},
		}

		internalServiceStructure.Childs = append(internalServiceStructure.Childs, moduleStructure)
		apiStructure.Skip = true
		apiProtoStructure.SkipIfExist, apiGraphQLStructure.SkipIfExist, apiJSONSchemaStructure.SkipIfExist = true, true, true
		apiStructure.Childs = []FileStructure{
			apiProtoStructure, apiGraphQLStructure, apiJSONSchemaStructure,
		}

		baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, []FileStructure{
			apiStructure, cmdStructure, internalServiceStructure, pkgServiceStructure,
		}...)
		baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, FileStructure{
			Source: srvConfig.toJSONString(), FileName: "candi.json",
		})
		baseDirectoryFile.Skip = true
		baseDirectoryFile.TargetDir = ""
		if flagParam.serviceName != "" && srvConfig.IsMonorepo {
			baseDirectoryFile.TargetDir = flagParam.outputFlag + flagParam.serviceName + "/"
		}
	}

	execGenerator(baseDirectoryFile)
	if srvConfig.GraphQLHandler {
		updateGraphQLRoot(&flagParam, srvConfig)
	}
	updateSharedUsecase(flagParam, srvConfig)
	updateSharedRepository(flagParam, srvConfig)
}

func monorepoGenerator(flagParam flagParameter) {
	var srvConfig serviceConfig
	srvConfig.Header = fmt.Sprintf("Code generated by candi %s.", candi.Version)
	srvConfig.IsMonorepo = true
	srvConfig.LibraryName = flagParam.libraryNameFlag
	sdkStructure := FileStructure{
		TargetDir: "sdk/", IsDir: true, Childs: []FileStructure{
			{FileName: ".gitkeep"},
			{FromTemplate: true, DataSource: srvConfig, Source: templateSDK, FileName: "sdk.go"},
		},
	}
	globalShared := FileStructure{
		TargetDir: "globalshared/", IsDir: true, Childs: []FileStructure{
			{FileName: ".gitkeep"},
		},
	}

	var baseDirectoryFile FileStructure
	baseDirectoryFile.TargetDir = flagParam.monorepoProjectName + "/"
	baseDirectoryFile.IsDir = true
	baseDirectoryFile.Childs = []FileStructure{
		globalShared,
		sdkStructure,
		{TargetDir: "services/", IsDir: true, Childs: []FileStructure{{FileName: ".gitkeep"}}},
		{FileName: "go.mod", Source: "module monorepo\n\ngo 1.16\n\nrequire github.com/golangid/candi " + candi.Version},
		{Source: gitignoreMonorepoTemplate, FileName: ".gitignore"},
		{Source: makefileMonorepoTemplate, FileName: "Makefile"},
		{Source: readmeMonorepoTemplate, FileName: "README.md"},
		{Source: dockerfileMonorepoTemplate, FileName: "Dockerfile"},
	}

	execGenerator(baseDirectoryFile)
}

func execGenerator(fl FileStructure) {
	if fl.Skip {
		goto execChild
	}

	if fl.SkipFunc != nil && fl.SkipFunc() {
		return
	}

	if fl.IsDir {
		if _, err := os.Stat(fl.TargetDir); os.IsExist(err) {
			if fl.SkipIfExist {
				goto execChild
			}
			log.Fatal(err)
		}

		if err := os.Mkdir(fl.TargetDir, 0700); err != nil {
			if fl.SkipIfExist {
				goto execChild
			}
			log.Fatal("mkdir err: ", err)
		}
	}

	if fl.FileName != "" {
		var buff []byte
		if fl.FromTemplate {
			if fl.Source != "" {
				buff = loadTemplate(fl.Source, fl.DataSource)
			} else {
				lastDir := filepath.Dir(fl.TargetDir)
				buff = defaultDataSource(lastDir[strings.LastIndex(lastDir, "/")+1:])
			}
		} else {
			buff = []byte(fl.Source)
		}
		if _, err := os.Stat(fl.TargetDir + fl.FileName); err == nil && fl.SkipIfExist {
			goto execChild
		}
		fmt.Printf("creating %s...\n", fl.TargetDir+fl.FileName)
		if err := os.WriteFile(fl.TargetDir+fl.FileName, buff, 0644); err != nil {
			log.Fatal(err)
		}
	}

execChild:
	for _, child := range fl.Childs {
		child.TargetDir = fl.TargetDir + child.TargetDir
		execGenerator(child)
	}
}
