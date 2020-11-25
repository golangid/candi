package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func main() {
	flag.StringVar(&scopeFlag, "scope", "", "set scope (1 for init service, 2 for add module(s)")
	flag.StringVar(&serviceNameFlag, "servicename", "", "define service name")
	flag.StringVar(&gomodName, "gomod", "", "define module prefix")
	flag.Parse()

	printBanner()

	scope, headerConfig, srvConfig, modConfigs, baseConfig := parseInput()

	if scope == addModule {
		files, err := ioutil.ReadDir("internal/modules")
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if f.IsDir() {
				modConfigs = append(modConfigs, moduleConfig{
					ModuleName: f.Name(), Skip: true,
				})
			}
		}
		if serviceNameFlag == "" {
			pwd, _ := os.Getwd()
			headerConfig.ServiceName = filepath.Base(pwd)
		}
	}

	sort.Slice(modConfigs, func(i, j int) bool {
		return modConfigs[i].ModuleName < modConfigs[j].ModuleName
	})

	srvConfig.configHeader = headerConfig
	srvConfig.config = baseConfig
	srvConfig.Modules = modConfigs
	srvConfigEdited := srvConfig

	tpl = template.New(packageName)

	apiStructure := FileStructure{
		TargetDir: "api/", IsDir: true,
	}
	cmdStructure := FileStructure{
		TargetDir: "cmd/", IsDir: true,
		Childs: []FileStructure{
			{TargetDir: "{{.ServiceName}}/", IsDir: true, DataSource: srvConfig},
		},
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
		TargetDir: "graphql/", IsDir: true,
	}

	moduleStructure := FileStructure{
		TargetDir: "modules/", IsDir: true, DataSource: srvConfig,
	}
	var sharedDomainFiles []FileStructure

	for i, mod := range srvConfig.Modules {
		mod.configHeader = srvConfig.configHeader
		mod.config = srvConfig.config
		if mod.Skip && scope == addModule {
			_, err := ioutil.ReadFile(fmt.Sprintf("internal/modules/%s/repository/repository_sql.go", mod.ModuleName))
			mod.SQLDeps = err == nil
			_, err = ioutil.ReadFile(fmt.Sprintf("internal/modules/%s/repository/repository_mongo.go", mod.ModuleName))
			mod.MongoDeps = err == nil
		}

		if mod.SQLDeps {
			srvConfigEdited.SQLDeps = mod.SQLDeps
		}
		if mod.MongoDeps {
			srvConfigEdited.MongoDeps = mod.MongoDeps
		}
		srvConfigEdited.Modules[i] = mod
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
					{TargetDir: "graphqlhandler/", IsDir: true,
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlRootTemplate, FileName: "root_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlQueryTemplate, FileName: "query_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlMutationTemplate, FileName: "mutation_resolver.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryGraphqlSubscriptionTemplate, FileName: "subscription_resolver.go"},
						}},
					{TargetDir: "grpchandler/", IsDir: true,
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryGRPCTemplate, FileName: "grpchandler.go"},
						}},
					{TargetDir: "resthandler/", IsDir: true,
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryRestTemplate, FileName: "resthandler.go"},
						}},
					{TargetDir: "workerhandler/", IsDir: true,
						Childs: []FileStructure{
							{FromTemplate: true, DataSource: mod, Source: deliveryKafkaTemplate, FileName: "kafka_handler.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryRedisTemplate, FileName: "redis_handler.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryCronTemplate, FileName: "cron_handler.go"},
							{FromTemplate: true, DataSource: mod, Source: deliveryTaskQueueTemplate, FileName: "taskqueue_handler.go"},
						}},
				},
			},
			{
				TargetDir: "domain/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, FileName: "payload.go"},
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
					{FromTemplate: true, DataSource: mod, Source: templateUsecaseImpl, FileName: "usecase_impl.go"},
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
			TargetDir: mod.ModuleName, IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, DataSource: mod, Source: defaultGRPCProto, FileName: mod.ModuleName + ".proto"},
			},
		})
		apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, FileStructure{
			FromTemplate: true, DataSource: mod, Source: defaultGraphqlSchema, FileName: mod.ModuleName + ".graphql",
		})

		// for shared domain
		sharedDomainFiles = append(sharedDomainFiles, FileStructure{
			Source: "package domain", FileName: mod.ModuleName + ".go",
		})
	}

	configsStructure := FileStructure{
		TargetDir: "configs/", IsDir: true,
		Childs: []FileStructure{
			{FromTemplate: true, DataSource: srvConfigEdited, Source: configsTemplate, FileName: "configs.go"},
		},
	}

	internalServiceStructure.Childs = append(internalServiceStructure.Childs, moduleStructure)
	internalServiceStructure.Childs = append(internalServiceStructure.Childs, FileStructure{
		FromTemplate: true, DataSource: srvConfig, Source: serviceMainTemplate, FileName: "service.go",
	})

	apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, []FileStructure{
		{FromTemplate: true, DataSource: srvConfig, Source: defaultGraphqlRootSchema, FileName: "_schema.graphql"},
	}...)
	apiStructure.Childs = []FileStructure{
		apiGraphQLStructure,
		{
			TargetDir: "jsonschema/", IsDir: true,
			Childs: []FileStructure{
				{DataSource: `{}`, FileName: "schema.json"},
			},
		},
		apiProtoStructure,
	}

	var baseDirectoryFile FileStructure
	baseDirectoryFile.TargetDir = "{{.ServiceName}}/"
	baseDirectoryFile.DataSource = srvConfig
	baseDirectoryFile.IsDir = true
	switch scope {
	case initService:
		pkgServiceStructure.Childs = append(pkgServiceStructure.Childs, []FileStructure{
			{TargetDir: "helper/", IsDir: true, Childs: []FileStructure{
				{FromTemplate: true, FileName: "helper.go"},
			}},
			{TargetDir: "shared/", IsDir: true, Childs: []FileStructure{
				{TargetDir: "domain/", IsDir: true, Childs: sharedDomainFiles},
				{TargetDir: "repository/", IsDir: true, Childs: parseSharedRepository(srvConfig)},
				{TargetDir: "usecase/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, DataSource: srvConfig, Source: templateUsecaseUOW, FileName: "usecase.go"},
				}},
				{FromTemplate: true, DataSource: srvConfig, Source: templateSharedTokenValidator, FileName: "token_validator.go"},
				{Source: additionalEnvTemplate, FromTemplate: true, DataSource: srvConfig, FileName: "environment.go"},
			}},
		}...)
		baseDirectoryFile.Childs = []FileStructure{
			apiStructure, cmdStructure, configsStructure, internalServiceStructure, pkgServiceStructure,
			{TargetDir: "deployments/", IsDir: true, Childs: []FileStructure{
				{TargetDir: "k8s/", IsDir: true, Childs: []FileStructure{
					{FileName: srvConfig.ServiceName + ".yml"},
				}},
			}},
			{TargetDir: "docs/", IsDir: true, Childs: []FileStructure{
				{FileName: ".gitkeep"},
			}},
			{Source: gitignoreTemplate, FileName: ".gitignore"},
			{FromTemplate: true, DataSource: srvConfig, Source: makefileTemplate, FileName: "Makefile"},
			{FromTemplate: true, DataSource: srvConfig, Source: dockerfileTemplate, FileName: "Dockerfile"},
			{FromTemplate: true, DataSource: srvConfig, Source: gomodTemplate, FileName: "go.mod"},
			{FromTemplate: true, DataSource: srvConfig, Source: cmdMainTemplate, FileName: "main.go"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env.sample"},
		}

	case addModule:
		configsStructure.Skip = true
		moduleStructure.Skip = true
		pkgServiceStructure.Skip = true
		pkgServiceStructure.Childs = []FileStructure{
			moduleStructure,
			{TargetDir: "shared/", IsDir: true, Skip: true, Childs: []FileStructure{
				{TargetDir: "domain/", IsDir: true, Skip: true, Childs: sharedDomainFiles},
				{TargetDir: "repository/", IsDir: true, Skip: true, Childs: parseSharedRepository(srvConfig)},
				{TargetDir: "usecase/", IsDir: true, Skip: true, Childs: []FileStructure{
					{FromTemplate: true, DataSource: srvConfig, Source: templateUsecaseUOW, FileName: "usecase.go"},
				}},
			}},
		}

		if srvConfig.SQLDeps {
			baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, FileStructure{
				FromTemplate: true, DataSource: srvConfig, Source: cmdMainTemplate, FileName: "main.go",
			})
		}

		apiStructure.Skip = true
		apiProtoStructure.Skip, apiGraphQLStructure.Skip = true, true
		apiStructure.Childs = []FileStructure{
			apiProtoStructure, apiGraphQLStructure,
		}

		baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, []FileStructure{
			apiStructure, configsStructure, pkgServiceStructure,
		}...)
		baseDirectoryFile.Skip = true
		baseDirectoryFile.TargetDir = ""

	}

	exec(baseDirectoryFile)
}

func exec(fl FileStructure) {
	dirBuff := loadTemplate(fl.TargetDir, fl.DataSource)
	dirName := string(dirBuff)

	if fl.Skip {
		goto execChild
	}

	if _, err := os.Stat(dirName); os.IsExist(err) {
		panic(err)
	}

	if fl.IsDir {
		fmt.Printf("creating %s...\n", dirName)
		if err := os.Mkdir(dirName, 0700); err != nil {
			fmt.Println("mkdir err:", err)
			panic(err)
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
		dirName = strings.TrimSuffix(dirName, "/")
		if err := ioutil.WriteFile(dirName+"/"+fl.FileName, buff, 0644); err != nil {
			panic(err)
		}
	}

execChild:
	for _, child := range fl.Childs {
		child.TargetDir = dirName + child.TargetDir
		exec(child)
	}
}
