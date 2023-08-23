package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/golangid/candi"
	"github.com/golangid/candi/candihelper"
)

func parseSharedRepository(data serviceConfig) (repos []FileStructure) {
	for i := range data.Modules {
		data.Modules[i].config = data.config
	}
	repos = append(repos, []FileStructure{
		{FromTemplate: true, DataSource: data, Source: templateRepository, FileName: "repository.go"},
		{FromTemplate: true, Skip: !data.SQLDeps, DataSource: data, Source: templateRepositoryUOWSQL, FileName: "repository_sql.go"},
		{FromTemplate: true, Skip: !data.MongoDeps, DataSource: data, Source: templateRepositoryUOWMongo, FileName: "repository_mongo.go"},
		{FromTemplate: true, Skip: !data.ArangoDeps, DataSource: data, Source: templateRepositoryUOWArango, FileName: "repository_arango.go"},
	}...)
	return
}

func parseRepositoryModule(data moduleConfig) (repos []FileStructure) {
	repos = append(repos, []FileStructure{
		{FromTemplate: true, Skip: !data.SQLDeps, DataSource: data, Source: templateRepositorySQLImpl,
			FileName: "repository_" + cleanSpecialChar.Replace(strings.ToLower(data.ModuleName)) + "_sql.go"},
		{FromTemplate: true, Skip: !data.MongoDeps, DataSource: data, Source: templateRepositoryMongoImpl,
			FileName: "repository_" + cleanSpecialChar.Replace(strings.ToLower(data.ModuleName)) + "_mongo.go"},
		{FromTemplate: true, Skip: !data.ArangoDeps, DataSource: data, Source: templateRepositoryArangoImpl,
			FileName: "repository_" + cleanSpecialChar.Replace(strings.ToLower(data.ModuleName)) + "_arango.go"},
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
		"camel": func(v string) string {
			return candihelper.ToCamelCase(v)
		},
		"snake": func(v string) string {
			return candihelper.ToDelimited(v, '_')
		},
		"plural": func(v string) string {
			return candihelper.ToDelimited(pluralize.NewClient().Plural(v), '_')
		},
		"kebab": func(v string) string {
			return candihelper.ToDelimited(v, '-')
		},
		"cleanPathModule": func(v string) string {
			return strings.ToLower(candihelper.ToDelimited(v, '-'))
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
	_, errSdk := os.ReadDir("sdk/")
	_, errService := os.ReadDir("services/")
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
		fmt.Printf(RedFormat, errMessage+", try again")
		serviceName = inputServiceName()
	}
	return
}

func inputOwnerName() (ownerName string) {
	ownerName = readInput("Please input owner name:")
	var errMessage string
	if strings.TrimSpace(ownerName) == "" {
		errMessage = "Owner name cannot empty"
	}

	if errMessage != "" {
		fmt.Printf(RedFormat, errMessage+", try again")
		ownerName = inputOwnerName()
	}
	return
}

func readInput(cmds ...string) string {
	logger.Printf("\033[1m%s\033[0m ", strings.Join(cmds, "\n"))
	fmt.Printf(">> ")
	cmdInput, _ := reader.ReadString('\n')
	return strings.TrimRight(strings.TrimSpace(cmdInput), "\n")
}

func validateDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf(RedFormat, fmt.Sprintf(`Directory "%s" is not exist`, dir))
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
	if flagParam.isMonorepo {
		baseDir = flagParam.outputFlag + flagParam.serviceName + "/"
	}

	b, err := os.ReadFile(baseDir + "candi.json")
	if err != nil {
		return serviceConfig{}
	}
	var savedConfig serviceConfig
	json.Unmarshal(b, &savedConfig)
	for i := range savedConfig.Modules {
		savedConfig.Modules[i].Skip = true
	}
	if err := checkVersion(candi.Version, savedConfig.Version); err != nil {
		log.Fatal(err)
	}
	savedConfig.Version = candi.Version
	savedConfig.IsMonorepo = flagParam.isMonorepo
	return savedConfig
}

func filterServerHandler(cfg *serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if !cfg.RestHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "resthandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) REST API", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RestHandler
	}
	if !cfg.GRPCHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "grpchandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) GRPC", len(options)+1))
		handlers[strconv.Itoa(len(options))] = GrpcHandler
	}
	if !cfg.GraphQLHandler || (flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "graphqlhandler")) != nil) {
		options = append(options, fmt.Sprintf("%d) GraphQL", len(options)+1))
		handlers[strconv.Itoa(len(options))] = GraphqlHandler
	}

	wording = strings.Join(options, "\n")
	return
}

func filterWorkerHandler(cfg *serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if !cfg.KafkaHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "kafka_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Kafka Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = KafkaHandler
	}
	if !cfg.SchedulerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "cron_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Scheduler", len(options)+1))
		handlers[strconv.Itoa(len(options))] = SchedulerHandler
	}
	if !cfg.RedisSubsHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "redis_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Redis Subscriber", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RedissubsHandler
	}
	if !cfg.TaskQueueHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "taskqueue_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Task Queue Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = TaskqueueHandler
	}
	if !cfg.PostgresListenerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "postgres_listener_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) Postgres Event Listener Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = PostgresListenerHandler
	}
	if !cfg.RabbitMQHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "rabbitmq_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) RabbitMQ Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RabbitmqHandler
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

func getDefaultPackageName() (packageName string) {
	packageOptions := strings.Split(os.Getenv(CandiPackagesEnv), ",")
	if len(packageOptions) == 1 && packageOptions[0] != "" {
		return packageOptions[0]
	}
	return DefaultPackageName
}

func getGoVersion() (version string) {
	version = strings.TrimPrefix(runtime.Version(), "go")
	if versionDetails := strings.Split(version, "."); len(versionDetails) > 2 {
		version = strings.Join(versionDetails[:len(versionDetails)-1], ".")
	}
	return
}

func checkVersion(cli, project string) error {
	cli, project = strings.Trim(cli, "v"), strings.Trim(project, "v")

	cliSplit := strings.Split(cli, ".")
	projectSplit := make([]int, len(cliSplit))

	for i, p := range strings.Split(project, ".") {
		if i >= len(projectSplit) {
			break
		}
		projectSplit[i], _ = strconv.Atoi(p)
	}

	for i, s := range cliSplit {
		c, _ := strconv.Atoi(s)
		if c < projectSplit[i] {
			return fmt.Errorf("ERROR: Your cli version (%s) must greater than candi version in service (%s), please upgrade your CLI", cli, project)
		} else if c > projectSplit[i] {
			return nil
		}
	}
	return nil
}
