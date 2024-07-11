package main

import (
	"log"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func addUsecase(flagParam *flagParameter, usecaseName string) {
	targetDir := ""
	if isWorkdirMonorepo() {
		targetDir = flagParam.outputFlag + flagParam.serviceName + "/"
	}

	fs := FileStructure{
		FromTemplate: true,
		DataSource:   map[string]any{"ModuleName": flagParam.moduleName, "UsecaseName": usecaseName},
		Source:       templateNewUsecase,
		FileName:     candihelper.ToDelimited(usecaseName, '_') + ".go",
	}
	if err := fs.writeFile(targetDir + "internal/modules/" + flagParam.moduleName + "/usecase"); err != nil {
		log.Fatal(err)
	}

	fu := fileUpdate{
		filepath:   targetDir + "internal/modules/" + flagParam.moduleName + "/usecase/usecase.go",
		oldContent: `type ` + strings.Title(candihelper.ToCamelCase(flagParam.moduleName)) + `Usecase interface {`,
		newContent: `type ` + strings.Title(candihelper.ToCamelCase(flagParam.moduleName)) + `Usecase interface {
	` + strings.Title(candihelper.ToCamelCase(usecaseName)) + `(ctx context.Context) (err error)`,
	}
	fu.readFileAndApply()
}
