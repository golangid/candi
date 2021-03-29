package main

import (
	"bufio"
	"flag"
	"fmt"
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
		"3 for init monorepo codebase, \n4 for init service in monorepo, \n5 for add module(s) service in monorepo, \n"+
		"6 for run multiple service in monorepo")
	flag.BoolVar(&flagParam.initService, "init", false, "[project generator] init service")
	flag.BoolVar(&flagParam.addModule, "add-module", false, "[project generator] add module in service")
	flag.BoolVar(&flagParam.initMonorepo, "init-monorepo", false, "[project generator] init monorepo codebase")
	flag.StringVar(&flagParam.packagePrefixFlag, "packageprefix", "", "[project generator] define package prefix")
	flag.BoolVar(&flagParam.withGoModFlag, "withgomod", true, "[project generator] generate go.mod or not")
	flag.StringVar(&flagParam.protoOutputPkgFlag, "protooutputpkg", "", "[project generator] define generated proto output target (if using grpc), with prefix is your go.mod")
	flag.StringVar(&flagParam.outputFlag, "output", "", "[project generator] directory to write project to (default is service name)")
	flag.StringVar(&flagParam.libraryNameFlag, "libraryname", "pkg.agungdp.dev/candi", "[project generator] define library name")

	flag.BoolVar(&flagParam.run, "run", false, "[service runner] run selected service or all service in monorepo")
	flag.StringVar(&flagParam.serviceName, "service", "", `Describe service name (if run multiple services, separate by comma)`)
	flag.Parse()

	tpl = template.New(flagParam.libraryNameFlag)

	switch {
	case flagParam.version:
		fmt.Printf("Build info:\n* Runtime version: %s\n", runtime.Version())

	case flagParam.run:
		serviceRunner(flagParam.serviceName)

	case flagParam.initMonorepo:
		monorepoGenerator(flagParam)

	case flagParam.initService:
		flagParam.scopeFlag = "1"
		if isWorkdirMonorepo() {
			flagParam.parseInitMonorepoService()
		}
		headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
		projectGenerator(flagParam, initService, headerConfig, srvConfig, modConfigs, baseConfig)

	case flagParam.addModule:
		flagParam.scopeFlag = "2"
		if isWorkdirMonorepo() {
			if flagParam.serviceName == "" {
				fmt.Printf(redFormat, "missing service name, make sure to include '-service' flag")
				return
			}
			flagParam.parseAddModuleMonorepoService()
		}
		headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
		projectGenerator(flagParam, addModule, headerConfig, srvConfig, modConfigs, baseConfig)

	default:
	selectScope:
		reader := bufio.NewReader(os.Stdin)
		if flagParam.scopeFlag == "" {
			fmt.Printf("\033[1mWhat do you want?\n" +
				"1) Init service\n" +
				"2) Add module(s)\n" +
				"3) Init monorepo codebase\n" +
				"4) Init service in monorepo\n" +
				"5) Add module(s) service in monorepo\n" +
				"6) Run multiple service in monorepo\033[0m\n>> ")
			cmdInput, _ := reader.ReadString('\n')
			cmdInput = strings.TrimRight(cmdInput, "\n")
			flagParam.scopeFlag = cmdInput
		}

		scope, ok := scopeMap[flagParam.scopeFlag]
		if !ok {
			fmt.Printf(redFormat, "Invalid scope option, please input valid scope below and try again")
			flagParam.scopeFlag = ""
			goto selectScope
		}
		selectScope(flagParam, scope)
	}
}

func selectScope(flagParam flagParameter, scope string) {
	switch scope {
	case initMonorepo: // 3
		monorepoGenerator(flagParam)
	case initMonorepoService: // 4
		flagParam.parseInitMonorepoService()
		headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
		projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
	case addModuleMonorepoService: // 5
		flagParam.parseAddModuleMonorepoService()
		headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
		projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
	case runServiceMonorepo: // 6
		serviceRunner(flagParam.serviceName)

	default:
		headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
		projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
	}
}
