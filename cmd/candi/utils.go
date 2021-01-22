package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"pkg.agungdwiprasetyo.com/candi/candihelper"
)

func parseInput() (scope string, headerConfig configHeader, srvConfig serviceConfig, modConfigs []moduleConfig, baseConfig config) {

	reader := bufio.NewReader(os.Stdin)
	serviceHandlers := make(map[string]bool)
	workerHandlers := make(map[string]bool)
	dependencies := make(map[string]bool)

selectScope:
	var cmdInput string
	if scopeFlag == "" {
		fmt.Println("\033[1mWhat do you want?\n" +
			"1) Init service\n" +
			"2) Add module(s)\033[0m")
		cmdInput, _ := reader.ReadString('\n')
		cmdInput = strings.TrimRight(cmdInput, "\n")
		scopeFlag = cmdInput
	}

	var ok bool
	scope, ok = scopeMap[scopeFlag]
	if !ok {
		fmt.Printf(redFormat, "Invalid option, try again")
		scopeFlag = ""
		goto selectScope
	}

	if scope == initService {
	inputServiceName:
		fmt.Print(ps1 + "\033[1mPlease input service name:\033[0m ")
		cmdInput, _ := reader.ReadString('\n')
		headerConfig.ServiceName = strings.TrimRight(cmdInput, "\n")
		_, err := os.Stat(headerConfig.ServiceName)
		var errMessage string
		if strings.TrimSpace(headerConfig.ServiceName) == "" {
			errMessage = "Service name cannot empty"
		}
		if !os.IsNotExist(err) {
			errMessage = "Folder already exists"
		}
		if errMessage != "" {
			fmt.Printf(redFormat, errMessage+", try again")
			cmdInput = ""
			goto inputServiceName
		}
	} else if scope == addModule && serviceNameFlag != "" {
		headerConfig.ServiceName = serviceNameFlag
	}

inputModules:
	fmt.Print(ps1 + "\033[1mPlease input module names (separated by comma):\033[0m ")
	cmdInput, _ = reader.ReadString('\n')
	cmdInput = strings.TrimRight(cmdInput, "\n")
	if strings.TrimSpace(cmdInput) == "" {
		fmt.Printf(redFormat, "Modules cannot empty")
		cmdInput = ""
		goto inputModules
	}
	for _, moduleName := range strings.Split(cmdInput, ",") {
		modConfigs = append(modConfigs, moduleConfig{
			ModuleName: strings.TrimSpace(moduleName), Skip: false,
		})
	}
	if scope == addModule {
		goto constructConfig
	}

selectServiceHandler:
	fmt.Print(ps1 + "\033[1mPlease select service handlers (separated by comma, enter for skip)\n" +
		"1) Rest API\n" +
		"2) GRPC\n" +
		"3) GraphQL\033[0m\n")
	cmdInput, _ = reader.ReadString('\n')
	cmdInput = strings.TrimRight(cmdInput, "\n")
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if serverName, ok := serviceHandlersMap[strings.TrimSpace(str)]; ok {
			serviceHandlers[serverName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			cmdInput = ""
			goto selectServiceHandler
		}
	}

selectWorkerHandlers:
	fmt.Print(ps1 + "\033[1mPlease select worker handlers (separated by comma, enter for skip)\n" +
		"1) Kafka Consumer\n" +
		"2) Scheduler\n" +
		"3) Redis Subscriber\n" +
		"4) Task Queue Worker\033[0m\n")
	cmdInput, _ = reader.ReadString('\n')
	cmdInput = strings.TrimRight(cmdInput, "\n")
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if workerName, ok := workerHandlersMap[strings.TrimSpace(str)]; ok {
			workerHandlers[workerName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			cmdInput = ""
			goto selectWorkerHandlers
		}
	}

	if len(serviceHandlers) == 0 && len(workerHandlers) == 0 {
		fmt.Printf(redFormat, "No service/worker selected, try again")
		cmdInput = ""
		goto selectServiceHandler
	}

selectDependencies:
	fmt.Print(ps1 + "\033[1mPlease select dependencies (separated by comma, enter for skip)\n" +
		"1) Redis\n" +
		"2) SQL Database\n" +
		"3) Mongo Database\033[0m\n")
	cmdInput, _ = reader.ReadString('\n')
	cmdInput = strings.TrimRight(cmdInput, "\n")

	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		str = strings.TrimSpace(str)
		if depsName, ok := dependencyMap[str]; ok {
			dependencies[depsName] = true
		} else if str != "" {
			fmt.Printf(redFormat, "Invalid option, try again")
			cmdInput = ""
			goto selectDependencies
		}
	}
	if workerHandlers[redissubsHandler] && !dependencies[redisDeps] {
		fmt.Printf(redFormat, "Redis Subscriber need redis, try again")
		cmdInput = ""
		goto selectDependencies
	}
	if workerHandlers[taskqueueHandler] && !(dependencies[redisDeps] && dependencies[mongodbDeps]) {
		fmt.Printf(redFormat, "Task Queue Worker need redis (for queue) and mongo (for log storage), try again")
		cmdInput = ""
		goto selectDependencies
	}

	if dependencies[sqldbDeps] {
	selectSQLDriver:
		fmt.Print(ps1 + "\033[1mPlease select SQL database driver (choose one)\n" +
			"1) Postgres\n" +
			"2) MySQL\033[0m\n")
		cmdInput, _ = reader.ReadString('\n')
		cmdInput = strings.TrimRight(strings.TrimSpace(cmdInput), "\n")
		baseConfig.SQLDriver, ok = sqlDrivers[cmdInput]
		if !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			cmdInput = ""
			goto selectSQLDriver
		}

	useGORMLabel:
		fmt.Print(ps1 + "\033[1mUse GORM? (y/n)\033[0m\n")
		cmdInput, _ = reader.ReadString('\n')
		cmdInput = strings.TrimRight(strings.TrimSpace(cmdInput), "\n")
		gormOpts := map[string]bool{"y": true, "n": false}
		if baseConfig.SQLUseGORM, ok = gormOpts[cmdInput]; !ok {
			fmt.Printf(redFormat, "Invalid option, try again")
			cmdInput = ""
			goto useGORMLabel
		}
	}

constructConfig:
	headerConfig.Header = fmt.Sprintf("Code generated by candi %s.", candihelper.Version)
	headerConfig.Version = candihelper.Version
	headerConfig.LibraryName = libraryNameFlag
	if packagePrefixFlag != "" {
		packagePrefixFlag = strings.TrimSuffix(packagePrefixFlag, "/") + "/"
		headerConfig.PackagePrefix = packagePrefixFlag + headerConfig.ServiceName
	} else {
		headerConfig.PackagePrefix = headerConfig.ServiceName
	}
	if protoOutputPkgFlag != "" {
		headerConfig.ProtoSource = protoOutputPkgFlag + "/" + headerConfig.ServiceName + "/proto"
	} else {
		headerConfig.ProtoSource = headerConfig.PackagePrefix + "/api/proto"
	}

	baseConfig.RestHandler = serviceHandlers[restHandler]
	baseConfig.GRPCHandler = serviceHandlers[grpcHandler]
	baseConfig.GraphQLHandler = serviceHandlers[graphqlHandler]
	baseConfig.KafkaHandler = workerHandlers[kafkaHandler]
	baseConfig.SchedulerHandler = workerHandlers[schedulerHandler]
	baseConfig.RedisSubsHandler = workerHandlers[redissubsHandler]
	baseConfig.TaskQueueHandler = workerHandlers[taskqueueHandler]
	baseConfig.RedisDeps = dependencies[redisDeps]
	baseConfig.SQLDeps, baseConfig.MongoDeps = dependencies[sqldbDeps], dependencies[mongodbDeps]
	baseConfig.IsWorkerActive = baseConfig.KafkaHandler ||
		baseConfig.SchedulerHandler ||
		baseConfig.RedisSubsHandler ||
		baseConfig.TaskQueueHandler

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
	replaceChar := []string{"*", "", "/", "", ":", ""}
	replacer := strings.NewReplacer(append(replaceChar, "-", "")...)
	modulePathReplacer := strings.NewReplacer(replaceChar...)
	return template.FuncMap{

		"clean": func(v string) string {
			return replacer.Replace(v)
		},
		"cleanPathModule": func(v string) string {
			return modulePathReplacer.Replace(v)
		},
		"upper": func(str string) string {
			return strings.Title(str)
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

`, candihelper.Version)
}
