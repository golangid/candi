package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func addUsecase(flagParam *flagParameter, usecaseName string, deliveryHandlers []string) {
	targetDir := ""
	packagePrefix := flagParam.serviceName
	if isWorkdirMonorepo() {
		targetDir = filepath.Join(flagParam.outputFlag, flagParam.serviceName) + "/"
		packagePrefix = filepath.Join(flagParam.packagePrefixFlag, flagParam.serviceName)
	}
	moduleDir := targetDir + "internal/modules/" + flagParam.moduleName

	newUsecase := FileStructure{
		FromTemplate: true, DataSource: map[string]any{
			"ModuleName": flagParam.moduleName, "UsecaseName": usecaseName,
			"PackagePrefix": packagePrefix,
		},
		Source:   templateNewUsecase,
		FileName: candihelper.ToDelimited(usecaseName, '_') + ".go",
	}
	if err := newUsecase.writeFile(moduleDir + "/usecase"); err != nil {
		log.Fatal(err)
	}

	modName := strings.Title(candihelper.ToCamelCase(flagParam.moduleName))
	ucName := strings.Title(candihelper.ToCamelCase(usecaseName))
	replaceFiles := []fileUpdate{
		{
			filepath:   moduleDir + "/usecase/usecase.go",
			oldContent: `type ` + modName + `Usecase interface {`,
			newContent: `type ` + modName + `Usecase interface {
	` + ucName + `(ctx context.Context, req *domain.Request` + ucName + `) (resp domain.Response` + ucName + `, err error)`,
		},
	}

	var fileUpdateContent []fileUpdate
	for _, dType := range []string{"Request", "Response"} {
		fileLoc := moduleDir + "/domain/" + strings.ToLower(dType) + ".go"
		if _, err := os.Stat(fileLoc); os.IsNotExist(err) {
			os.WriteFile(fileLoc, []byte("package domain\n\n"), 0644)
		}
		fileUpdateContent = append(fileUpdateContent, fileUpdate{
			filepath: fileLoc,
			newContent: string(`// ` + dType + ucName + ` model
type ` + dType + ucName + ` struct {
}
`),
		})
	}

	replacedDelivery, updatedDelivery := getUpdateFileInExistingDelivery(flagParam, usecaseName, deliveryHandlers)
	replaceFiles = append(replaceFiles, replacedDelivery...)
	fileUpdateContent = append(fileUpdateContent, updatedDelivery...)

	for _, fu := range fileUpdateContent {
		fu.addContent()
	}
	for _, fu := range replaceFiles {
		fu.readFileAndApply()
	}
}

func applyUsecaseToDelivery(flagParam *flagParameter, usecaseName string, deliveryHandlers []string) {
	replacedDelivery, updatedDelivery := getUpdateFileInExistingDelivery(flagParam, usecaseName, deliveryHandlers)
	for _, fu := range updatedDelivery {
		fu.addContent()
	}
	for _, fu := range replacedDelivery {
		fu.readFileAndApply()
	}
}

func getUpdateFileInExistingDelivery(flagParam *flagParameter, usecaseName string, deliveryHandlers []string) (replaceFiles, fileUpdateContent []fileUpdate) {
	targetDir := ""
	packagePrefix := flagParam.serviceName
	if isWorkdirMonorepo() {
		targetDir = filepath.Join(flagParam.outputFlag, flagParam.serviceName) + "/"
		packagePrefix = filepath.Join(flagParam.packagePrefixFlag, flagParam.serviceName)
	}
	moduleDir := targetDir + "internal/modules/" + flagParam.moduleName
	modName := strings.Title(candihelper.ToCamelCase(flagParam.moduleName))
	ucName := strings.Title(candihelper.ToCamelCase(usecaseName))

	mapWorkerDelivery := map[string]string{
		KafkaHandler:            "Kafka",
		SchedulerHandler:        "Cron",
		RedissubsHandler:        "Redis",
		TaskqueueHandler:        "TaskQueue",
		PostgresListenerHandler: "PostgresListener",
		RabbitmqHandler:         "RabbitMQ",
		pluginGCPPubSubWorker:   pluginGCPPubSubWorker,
		pluginSTOMPWorker:       pluginSTOMPWorker,
	}
	for _, delivery := range deliveryHandlers {
		fu := fileUpdate{
			filepath: moduleDir + "/delivery/" + deliveryHandlerLocation[delivery],
		}
		switch delivery {
		case RestHandler:
			fu.newContent = getRestFuncTemplate(modName, ucName)
			fu.skipContains = "func (h *RestHandler) " + candihelper.ToCamelCase(ucName)
			replaceFiles = append(replaceFiles, []fileUpdate{
				{filepath: fu.filepath, oldContent: `Mount(root interfaces.RESTRouter) {`,
					newContent: `Mount(root interfaces.RESTRouter) {
	root.GET(candihelper.V1+"/` + strings.ToLower(modName) + `/` + candihelper.ToDelimited(usecaseName, '-') + `", h.` + candihelper.ToCamelCase(usecaseName) + `)`,
					skipContains: `root.GET(candihelper.V1+"/` + strings.ToLower(modName) + `/` + candihelper.ToDelimited(usecaseName, '-') + `", h.` + candihelper.ToCamelCase(usecaseName) + `)`},
			}...)

		case GrpcHandler:
			fu.newContent = getGRPCFuncTemplate(modName, ucName)
			fu.skipContains = "func (h *GRPCHandler) " + ucName
			protoSource := targetDir + "api/proto/" + flagParam.moduleName + "/" + flagParam.moduleName + ".proto"
			replaceFiles = append(replaceFiles, []fileUpdate{
				{filepath: protoSource,
					oldContent: `service ` + modName + `Handler {`,
					newContent: `service ` + modName + `Handler {
	rpc ` + ucName + `(Request` + ucName + `) returns (Response` + ucName + `);`,
					skipContains: `rpc ` + ucName + `(Request` + ucName + `) returns (Response` + ucName + `);`},
			}...)
			fileUpdateContent = append(fileUpdateContent, fileUpdate{
				filepath: protoSource, newContent: `message Request` + ucName + ` {
}

message Response` + ucName + ` {
}
`,
				skipContains: `rpc ` + ucName + `(Request` + ucName + `) returns (Response` + ucName + `);`,
			})

		case GraphqlHandler:
			fu.newContent = getGraphQLFuncTemplate(modName, ucName)
			fu.skipContains = "func (m *GraphQLHandler) " + ucName
			gqlSchemaSource := targetDir + "api/graphql/" + flagParam.moduleName + ".graphql"
			replaceFiles = append(replaceFiles, []fileUpdate{
				{filepath: gqlSchemaSource, oldContent: `type ` + modName + `QueryResolver {`,
					newContent: `type ` + modName + `QueryResolver {
	` + candihelper.ToCamelCase(ucName) + `(data: ` + ucName + `InputResolver!): ` + ucName + `Resolver!`,
					skipContains: candihelper.ToCamelCase(ucName) + `(data: ` + ucName + `InputResolver!): ` + ucName + `Resolver!`},
			}...)
			fileUpdateContent = append(fileUpdateContent, fileUpdate{
				filepath: gqlSchemaSource, newContent: `input ` + ucName + `InputResolver {
}

type ` + ucName + `Resolver {
}
`,
				skipContains: candihelper.ToCamelCase(ucName) + `(data: ` + ucName + `InputResolver!): ` + ucName + `Resolver!`,
			})

		default:
			workerName, ok := mapWorkerDelivery[delivery]
			if !ok {
				continue
			}
			fu.newContent = getWorkerFuncTemplate(workerName, modName, ucName)
			fu.skipContains = "func (h *" + workerName + "Handler) " + candihelper.ToCamelCase(ucName)
			handlerRoutePattern := `"` + candihelper.ToDelimited(usecaseName, '-') + `"`
			if delivery == SchedulerHandler {
				handlerRoutePattern = `cronworker.CreateCronJobKey(` + handlerRoutePattern + `, "message", "* * * * *")`
			}

			replaceFiles = append(replaceFiles, []fileUpdate{
				{filepath: fu.filepath, oldContent: "\n	\"encoding/json\"", newContent: ""},
				{filepath: fu.filepath, oldContent: `import (`, newContent: `import (
	"encoding/json"`},
				{filepath: fu.filepath, oldContent: "\n	\"" + packagePrefix + "/internal/modules/" + flagParam.moduleName + "/domain\"", newContent: ""},
				{filepath: fu.filepath, oldContent: `import (`, newContent: `import (
	"` + packagePrefix + `/internal/modules/` + flagParam.moduleName + `/domain"`},
				{filepath: fu.filepath, oldContent: `MountHandlers(group *types.WorkerHandlerGroup) {`,
					newContent: `MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(` + handlerRoutePattern + `, h.` + candihelper.ToCamelCase(usecaseName) + `)`,
					skipContains: `group.Add(` + handlerRoutePattern + `, h.` + candihelper.ToCamelCase(usecaseName) + `)`},
			}...)
		}
		fileUpdateContent = append(fileUpdateContent, fu)
	}
	return
}
