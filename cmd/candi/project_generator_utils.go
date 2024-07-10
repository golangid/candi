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
	if len(cmds) > 0 {
		logger.Printf("\033[1m%s\033[0m ", strings.Join(cmds, "\n"))
		fmt.Printf(">> ")
	}
	cmdInput, _ := reader.ReadString('\n')
	return strings.TrimSpace(cmdInput)
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
		savedConfig.Modules[i].config = savedConfig.config
		savedConfig.Modules[i].configHeader = savedConfig.configHeader
	}
	if err := checkVersion(candi.Version, savedConfig.Version); err != nil {
		log.Fatal(err)
	}
	savedConfig.Version = candi.Version
	savedConfig.IsMonorepo = flagParam.isMonorepo
	savedConfig.OutputDir = flagParam.outputFlag
	return savedConfig
}

func filterServerHandler(cfg *serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if flagParam.addModule || (!cfg.RestHandler ||
		(flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "resthandler")) != nil)) {
		options = append(options, fmt.Sprintf("%d) REST API", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RestHandler
	}
	if flagParam.addModule || (!cfg.GRPCHandler ||
		(flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "grpchandler")) != nil)) {
		options = append(options, fmt.Sprintf("%d) GRPC", len(options)+1))
		handlers[strconv.Itoa(len(options))] = GrpcHandler
	}
	if flagParam.addModule || (!cfg.GraphQLHandler ||
		(flagParam.addHandler && validateDir(flagParam.getFullModuleChildDir("delivery", "graphqlhandler")) != nil)) {
		options = append(options, fmt.Sprintf("%d) GraphQL", len(options)+1))
		handlers[strconv.Itoa(len(options))] = GraphqlHandler
	}

	wording = strings.Join(options, "\n")
	return
}

func filterWorkerHandler(cfg *serviceConfig, flagParam *flagParameter) (wording string, handlers map[string]string) {
	handlers = make(map[string]string)
	var options []string
	if flagParam.addModule || (!cfg.KafkaHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "kafka_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) Kafka Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = KafkaHandler
	}
	if flagParam.addModule || (!cfg.SchedulerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "cron_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) Cron Scheduler", len(options)+1))
		handlers[strconv.Itoa(len(options))] = SchedulerHandler
	}
	if flagParam.addModule || (!cfg.RedisSubsHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "redis_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) Redis Subscriber", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RedissubsHandler
	}
	if flagParam.addModule || (!cfg.TaskQueueHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "taskqueue_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) Task Queue Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = TaskqueueHandler
	}
	if flagParam.addModule || (!cfg.PostgresListenerHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "postgres_listener_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) Postgres Event Listener Worker", len(options)+1))
		handlers[strconv.Itoa(len(options))] = PostgresListenerHandler
	}
	if flagParam.addModule || (!cfg.RabbitMQHandler || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", "rabbitmq_handler.go")) != nil)) {
		options = append(options, fmt.Sprintf("%d) RabbitMQ Consumer", len(options)+1))
		handlers[strconv.Itoa(len(options))] = RabbitmqHandler
	}
	if flagParam.addModule || flagParam.initService || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", strings.ToLower(pluginGCPPubSubWorker)+"_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) GCP PubSub Subscriber (plugin)", len(options)+1))
		handlers[strconv.Itoa(len(options))] = pluginGCPPubSubWorker
	}
	if flagParam.addModule || flagParam.initService || (flagParam.addHandler &&
		validateDir(flagParam.getFullModuleChildDir("delivery", "workerhandler", strings.ToLower(pluginSTOMPWorker)+"_handler.go")) != nil) {
		options = append(options, fmt.Sprintf("%d) AMQ (STOMP) Consumer (plugin)", len(options)+1))
		handlers[strconv.Itoa(len(options))] = pluginSTOMPWorker
	}

	wording = strings.Join(options, "\n")
	return
}

func getNeedFileUpdates(srvConfig *serviceConfig) (fileUpdates []fileUpdate) {
	rootDir := srvConfig.getRootDir()
	if srvConfig.RestHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_REST=false", newContent: "USE_REST=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_REST=false", newContent: "USE_REST=true"},
		}...)
	}
	if srvConfig.GRPCHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_GRPC=false", newContent: "USE_GRPC=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_GRPC=false", newContent: "USE_GRPC=true"},
		}...)
	}
	if srvConfig.GraphQLHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_GRAPHQL=false", newContent: "USE_GRAPHQL=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_GRAPHQL=false", newContent: "USE_GRAPHQL=true"},
			{filepath: rootDir + "api/api.go", oldContent: "// //go:embed all:graphql", newContent: "//go:embed all:graphql"},
		}...)
	}

	if srvConfig.KafkaHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_KAFKA_CONSUMER=false", newContent: "USE_KAFKA_CONSUMER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_KAFKA_CONSUMER=false", newContent: "USE_KAFKA_CONSUMER=true"},
			{filepath: rootDir + "configs/configs.go", oldContent: "// broker.NewKafkaBroker(),", newContent: "broker.NewKafkaBroker(),"},
		}...)
	}
	if srvConfig.SchedulerHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_CRON_SCHEDULER=false", newContent: "USE_CRON_SCHEDULER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_CRON_SCHEDULER=false", newContent: "USE_CRON_SCHEDULER=true"},
		}...)
	}
	if srvConfig.RedisSubsHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_REDIS_SUBSCRIBER=false", newContent: "USE_REDIS_SUBSCRIBER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_REDIS_SUBSCRIBER=false", newContent: "USE_REDIS_SUBSCRIBER=true"},
			{filepath: rootDir + "configs/configs.go", oldContent: "// broker.NewRedisBroker(redisDeps.WritePool()),", newContent: "broker.NewRedisBroker(redisDeps.WritePool()),"},
		}...)
	}
	if srvConfig.TaskQueueHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_TASK_QUEUE_WORKER=false", newContent: "USE_TASK_QUEUE_WORKER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_TASK_QUEUE_WORKER=false", newContent: "USE_TASK_QUEUE_WORKER=true"},
		}...)
	}
	if srvConfig.PostgresListenerHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_POSTGRES_LISTENER_WORKER=false", newContent: "USE_POSTGRES_LISTENER_WORKER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_POSTGRES_LISTENER_WORKER=false", newContent: "USE_POSTGRES_LISTENER_WORKER=true"},
		}...)
	}
	if srvConfig.RabbitMQHandler {
		fileUpdates = append(fileUpdates, []fileUpdate{
			{filepath: rootDir + ".env", oldContent: "USE_RABBITMQ_CONSUMER=false", newContent: "USE_RABBITMQ_CONSUMER=true"},
			{filepath: rootDir + ".env.sample", oldContent: "USE_RABBITMQ_CONSUMER=false", newContent: "USE_RABBITMQ_CONSUMER=true"},
			{filepath: rootDir + "configs/configs.go", oldContent: "// broker.NewRabbitMQBroker(),", newContent: "broker.NewRabbitMQBroker(),"},
		}...)
	}

	for _, module := range srvConfig.Modules {
		if module.Skip && !srvConfig.flag.addHandler {
			continue
		}
		moduleName := cleanSpecialChar.Replace(module.ModuleName)
		deliveryPackageDir := fmt.Sprintf(`"%s/internal/modules/%s/delivery`, module.PackagePrefix, moduleName)

		if module.RestHandler {
			fileUpdates = append(fileUpdates, []fileUpdate{
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// mod.restHandler", newContent: "mod.restHandler"},
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// " + deliveryPackageDir + "/resthandler", newContent: deliveryPackageDir + "/resthandler"},
			}...)
		}
		if module.GRPCHandler {
			fileUpdates = append(fileUpdates, []fileUpdate{
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// mod.grpcHandler", newContent: "mod.grpcHandler"},
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// " + deliveryPackageDir + "/grpchandler", newContent: deliveryPackageDir + "/grpchandler"},
			}...)
		}
		if module.GraphQLHandler {
			cleanMod, cleanUpperMod := candihelper.ToCamelCase(module.ModuleName), strings.Title(candihelper.ToCamelCase(module.ModuleName))
			fileUpdates = append(fileUpdates, []fileUpdate{
				{filepath: rootDir + "api/graphql/_schema.graphql",
					oldContent: "type Query {", newContent: fmt.Sprintf("type Query {\n	%s: %sQueryResolver @auth(authType: BEARER)", cleanMod, cleanUpperMod)},
				{filepath: rootDir + "api/graphql/_schema.graphql",
					oldContent: "type Mutation {", newContent: fmt.Sprintf("type Mutation {\n	%s: %sMutationResolver @auth(authType: BEARER)", cleanMod, cleanUpperMod)},
				{filepath: rootDir + "api/graphql/_schema.graphql",
					oldContent: "type Subscription {", newContent: fmt.Sprintf("type Subscription {\n	%s: %sSubscriptionResolver", cleanMod, cleanUpperMod)},
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// mod.graphqlHandler", newContent: "mod.graphqlHandler"},
				{filepath: rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// " + deliveryPackageDir + "/graphqlhandler", newContent: deliveryPackageDir + "/graphqlhandler"},
			}...)
		}
		if workerActivations := module.constructModuleWorkerActivation(); len(workerActivations) > 0 {
			fileUpdates = append(fileUpdates, fileUpdate{
				filepath:   rootDir + "internal/modules/" + moduleName + "/module.go",
				oldContent: "// " + deliveryPackageDir + "/workerhandler", newContent: deliveryPackageDir + "/workerhandler",
			})
			for _, workerActivation := range workerActivations {
				fileUpdates = append(fileUpdates, fileUpdate{
					filepath:   rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// " + workerActivation, newContent: workerActivation,
				})
			}
		}
	}

	for _, pl := range srvConfig.workerPlugins {
		for before, after := range pl.editConfig {
			fileUpdates = append(fileUpdates, fileUpdate{
				filepath: rootDir + "configs/configs.go", oldContent: before, newContent: after,
			})
		}
		for before, after := range pl.editAppFactory {
			fileUpdates = append(fileUpdates, fileUpdate{
				filepath: rootDir + "configs/app_factory.go", oldContent: before, newContent: after,
			})
		}
		for _, module := range srvConfig.Modules {
			if module.Skip {
				continue
			}
			moduleName := cleanSpecialChar.Replace(module.ModuleName)
			deliveryPackageDir := fmt.Sprintf(`"%s/internal/modules/%s/delivery`, module.PackagePrefix, moduleName)
			if !module.IsWorkerActive {
				fileUpdates = append(fileUpdates, fileUpdate{
					filepath:   rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: "// " + deliveryPackageDir + "/workerhandler",
					newContent: deliveryPackageDir + "/workerhandler",
				})
			}
			for before, after := range pl.editModule {
				fileUpdates = append(fileUpdates, fileUpdate{
					filepath:   rootDir + "internal/modules/" + moduleName + "/module.go",
					oldContent: before, newContent: after,
				})
			}
		}
	}
	return fileUpdates
}

type fileUpdate struct {
	filepath   string
	oldContent string
	newContent string
}

func (f *fileUpdate) readFileAndApply() {
	b, err := os.ReadFile(f.filepath)
	if err != nil {
		return
	}
	os.WriteFile(f.filepath, bytes.Replace(b, []byte(f.oldContent), []byte(f.newContent), -1), 0644)
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
