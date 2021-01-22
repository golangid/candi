package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func main() {
	flag.StringVar(&scopeFlag, "scope", "", "set scope (1 for init service, 2 for add module(s)")
	flag.StringVar(&serviceNameFlag, "servicename", "", "define service name")
	flag.StringVar(&packagePrefixFlag, "packageprefix", "", "define package prefix")
	flag.BoolVar(&withGoModFlag, "withgomod", true, "generate go.mod or not")
	flag.StringVar(&protoOutputPkgFlag, "protooutputpkg", "", "define generated proto output target (if using grpc), with prefix is your go.mod")
	flag.StringVar(&outputFlag, "output", "", "directory to write project to (default is service name)")
	flag.StringVar(&libraryNameFlag, "libraryname", "pkg.agungdwiprasetyo.com/candi", "define library name")
	flag.Parse()

	printBanner()

	scope, headerConfig, srvConfig, modConfigs, baseConfig := parseInput()

	srvConfig.configHeader = headerConfig
	srvConfig.config = baseConfig
	if scope == addModule {
		var baseDir string
		if serviceNameFlag != "" {
			baseDir = outputFlag + serviceNameFlag + "/"
		}

		b, err := ioutil.ReadFile(baseDir + "candi.json")
		if err != nil {
			log.Fatal("ERROR: cannot find candi.json file")
		}
		json.Unmarshal(b, &srvConfig)
		for i := range srvConfig.Modules {
			srvConfig.Modules[i].Skip = true
		}
		modConfigs = append(modConfigs, srvConfig.Modules...)
	}

	sort.Slice(modConfigs, func(i, j int) bool {
		return modConfigs[i].ModuleName < modConfigs[j].ModuleName
	})
	srvConfig.Modules = modConfigs

	tpl = template.New(libraryNameFlag)

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
			{FromTemplate: true, DataSource: srvConfig, Source: configsTemplate, FileName: "configs.go"},
		},
	}

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
				{Source: jsonSchemaTemplate, FromTemplate: true, FileName: "schema.json"},
			},
		},
		apiProtoStructure,
	}

	configJSON, _ := json.Marshal(srvConfig)

	var baseDirectoryFile FileStructure
	baseDirectoryFile.TargetDir = outputFlag + "{{.ServiceName}}/"
	baseDirectoryFile.DataSource = srvConfig
	baseDirectoryFile.IsDir = true
	switch scope {
	case initService:
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
			{FromTemplate: true, DataSource: srvConfig, Source: gitignoreTemplate, FileName: ".gitignore"},
			{FromTemplate: true, DataSource: srvConfig, Source: makefileTemplate, FileName: "Makefile"},
			{FromTemplate: true, DataSource: srvConfig, Source: dockerfileTemplate, FileName: "Dockerfile"},
			{FromTemplate: true, DataSource: srvConfig, Source: cmdMainTemplate, FileName: "main.go"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env"},
			{FromTemplate: true, DataSource: srvConfig, Source: envTemplate, FileName: ".env.sample"},
			{Source: string(configJSON), FileName: "candi.json"},
			{FromTemplate: true, DataSource: srvConfig, Source: readmeTemplate, FileName: "README.md"},
		}
		if withGoModFlag {
			baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, FileStructure{
				FromTemplate: true, DataSource: srvConfig, Source: gomodTemplate, FileName: "go.mod",
			})
		}

	case addModule:
		configsStructure.Skip = true
		moduleStructure.Skip = true
		pkgServiceStructure.Skip = true
		internalServiceStructure.Skip = true
		pkgServiceStructure.Childs = []FileStructure{
			{TargetDir: "shared/", IsDir: true, Skip: true, Childs: []FileStructure{
				{TargetDir: "domain/", IsDir: true, Skip: true, Childs: sharedDomainFiles},
				{TargetDir: "repository/", IsDir: true, Skip: true, Childs: parseSharedRepository(srvConfig)},
				{TargetDir: "usecase/", IsDir: true, Skip: true, Childs: []FileStructure{
					{FromTemplate: true, DataSource: srvConfig, Source: templateUsecaseUOW, FileName: "usecase.go"},
				}},
			}},
		}

		internalServiceStructure.Childs = append(internalServiceStructure.Childs, moduleStructure)
		apiStructure.Skip = true
		apiProtoStructure.Skip, apiGraphQLStructure.Skip = true, true
		apiStructure.Childs = []FileStructure{
			apiProtoStructure, apiGraphQLStructure,
		}

		baseDirectoryFile.Childs = append(baseDirectoryFile.Childs, []FileStructure{
			apiStructure, internalServiceStructure, pkgServiceStructure,
		}...)
		baseDirectoryFile.Skip = true
		baseDirectoryFile.TargetDir = ""
		if serviceNameFlag != "" {
			baseDirectoryFile.TargetDir = outputFlag + serviceNameFlag + "/"
		}

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
