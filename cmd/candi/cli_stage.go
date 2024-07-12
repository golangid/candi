package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golangid/candi/candihelper"
)

func cliStageInputServiceName() (serviceName string) {
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
		serviceName = cliStageInputServiceName()
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

func cliStageInputExistingServiceName(prefixPath, cmd string) string {
stageInputServiceName:
	serviceName := readInput(cmd)
	_, err := os.Stat(filepath.Join(prefixPath, serviceName))
	var errMessage string
	if strings.TrimSpace(serviceName) == "" {
		errMessage = "Service name cannot empty"
	}
	if os.IsNotExist(err) {
		errMessage = fmt.Sprintf(`Service "%s" is not exist in "%s" directory`, serviceName, prefixPath)
	}
	if errMessage != "" {
		fmt.Printf(RedFormat, errMessage+", try again")
		goto stageInputServiceName
	}
	return serviceName
}

func cliStageInputModule(prefixPath, cmd string) string {
stageReadInputModule:
	moduleName := candihelper.ToDelimited(readInput(cmd), '-')
	if err := validateDir(filepath.Join(prefixPath, moduleName)); err != nil {
		fmt.Print(err.Error())
		goto stageReadInputModule
	}
	return moduleName
}

func cliStageInputUsecaseName(prefixPath string) string {
stageAddUsecaseInputName:
	ucName := readInput("Please input usecase name:")
	dir := filepath.Join(prefixPath, "usecase", candihelper.ToDelimited(ucName, '_')+".go")
	if err := validateDir(dir); err == nil {
		fmt.Printf(RedFormat, "Usecase "+dir+" is exist, try another usecase name")
		goto stageAddUsecaseInputName
	}
	return ucName
}

func cliStageInputExistingUsecaseName(prefixPath string) string {
stageAddUsecaseInputName:
	ucName := readInput("Please input usecase name:")
	dir := filepath.Join(prefixPath, "usecase", candihelper.ToDelimited(ucName, '_')+".go")
	if err := validateDir(dir); err != nil {
		fmt.Printf(RedFormat, "Usecase "+dir+" not exist")
		goto stageAddUsecaseInputName
	}
	return ucName
}

func cliStageInputExistingDelivery(prefixPath string) (deliveryHandlers []string) {
	wording, deliveryHandlerMap := getAllModuleHandler(prefixPath)
stageSelectDeliveryHandler:
	cmdInput := readInput("Apply in delivery handler (separated by comma, enter for skip):\n" + wording)
	for _, str := range strings.Split(strings.Trim(cmdInput, ","), ",") {
		if delName, ok := deliveryHandlerMap[strings.TrimSpace(str)]; ok {
			deliveryHandlers = append(deliveryHandlers, delName)
		} else if str != "" {
			fmt.Printf(RedFormat, "Invalid option, try again")
			deliveryHandlers = []string{}
			goto stageSelectDeliveryHandler
		}
	}
	return deliveryHandlers
}
