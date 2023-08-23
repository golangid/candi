package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"text/template"
)

func main() {
	printBanner()

	var flagParam flagParameter
	flag.BoolVar(&flagParam.version, "version", false, "print version")
	flag.StringVar(&flagParam.scopeFlag, "scope", "", "[project generator] set scope \n1 for init service, \n2 for add module(s), \n"+
		"3 for add delivery handler(s) in module, \n4 for init monorepo codebase, \n5 for run multiple service in monorepo")
	flag.BoolVar(&flagParam.initService, "init", false, "[project generator] init service")
	flag.BoolVar(&flagParam.addModule, "add-module", false, "[project generator] add module in service")
	flag.BoolVar(&flagParam.addHandler, "add-handler", false, "[project generator] add handler in delivery module in service")
	flag.BoolVar(&flagParam.initMonorepo, "init-monorepo", false, "[project generator] init monorepo codebase")
	flag.StringVar(&flagParam.monorepoProjectName, "monorepo-name", "monorepo", "[project generator] set monorepo project name (default 'monorepo')")
	flag.StringVar(&flagParam.packagePrefixFlag, "packageprefix", "", "[project generator] define package prefix")
	flag.BoolVar(&flagParam.withGoModFlag, "withgomod", true, "[project generator] generate go.mod or not")
	flag.StringVar(&flagParam.protoOutputPkgFlag, "protooutputpkg", "", "[project generator] define generated proto output target (if using grpc), with prefix is your go.mod")
	flag.StringVar(&flagParam.outputFlag, "output", "", "[project generator] directory to write project to (default is service name)")
	flag.StringVar(&flagParam.libraryNameFlag, "libraryname", getDefaultPackageName(), "[project generator] define library name")

	flag.BoolVar(&flagParam.run, "run", false, "[service runner] run selected service or all service in monorepo")
	flag.StringVar(&flagParam.serviceName, "service", "", `Describe service name (if run multiple services, separate by comma)`)
	flag.Parse()

	tpl = template.New(flagParam.libraryNameFlag)
	logger = log.New(os.Stdout, "\x1b[32;1m[project generator]: \x1b[0m", log.Lmsgprefix)
	reader = bufio.NewReader(os.Stdin)
	flagParam.isMonorepo = isWorkdirMonorepo()

	switch {
	case flagParam.version:
		fmt.Printf("Build info:\n* Runtime version: %s\n", runtime.Version())

	case flagParam.run:
		serviceRunner(flagParam.serviceName)

	case flagParam.initMonorepo:
		monorepoGenerator(flagParam)

	case flagParam.initService:
		flagParam.scopeFlag = "2"
		if flagParam.isMonorepo {
			flagParam.parseMonorepoFlag()
		}
		projectGenerator(flagParam, InitService, parseInput(&flagParam))

	case flagParam.addModule:
		flagParam.scopeFlag = "3"
		if flagParam.isMonorepo {
			if err := flagParam.parseMonorepoFlag(); err != nil {
				fmt.Print(err.Error())
				return
			}
		}
		projectGenerator(flagParam, AddModule, parseInput(&flagParam))

	case flagParam.addHandler:
		flagParam.scopeFlag = "4"
		if isWorkdirMonorepo() {
			if err := flagParam.parseMonorepoFlag(); err != nil {
				fmt.Print(err.Error())
				return
			}
		}
		projectGenerator(flagParam, AddModule, parseInput(&flagParam))

	default:
	selectScope:
		if flagParam.scopeFlag == "" {
			fmt.Printf("\033[1mWhat do you want?\n" +
				"1) Init monorepo codebase\n" +
				"2) Init service\n" +
				"3) Add module(s) in service\n" +
				"4) Add delivery handler(s) in module\033[0m\n>> ")
			cmdInput, _ := reader.ReadString('\n')
			cmdInput = strings.TrimRight(cmdInput, "\n")
			flagParam.scopeFlag = cmdInput
		}

		scope, ok := scopeMap[flagParam.scopeFlag]
		if !ok {
			fmt.Printf(RedFormat, "Invalid scope option, please input valid scope below and try again")
			flagParam.scopeFlag = ""
			goto selectScope
		}
		selectScope(flagParam, scope)
		return
	}

}

func selectScope(flagParam flagParameter, scope string) {
	switch scope {
	case InitMonorepo: // 1
		logger.Printf("\033[1mPlease input monorepo project name (enter for default):\033[0m")
		fmt.Printf(">> ")
		if cmdInput, _ := reader.ReadString('\n'); strings.TrimRight(cmdInput, "\n") != "" {
			flagParam.monorepoProjectName = strings.TrimRight(cmdInput, "\n")
		}
		monorepoGenerator(flagParam)
		return
	case RunServiceMonorepo: // 5
		serviceRunner(flagParam.serviceName)
		return
	case InitService:
		flagParam.initService = true
	}

	if flagParam.isMonorepo {
		flagParam.parseMonorepoFlag()
	}
	projectGenerator(flagParam, scope, parseInput(&flagParam))
}
