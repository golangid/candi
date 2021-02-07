package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func main() {
	printBanner()

	var flagParam flagParameter
	flag.StringVar(&flagParam.scopeFlag, "scope", "", "[project generator] set scope \n1 for init service, \n2 for add module(s), \n"+
		"3 for init monorepo codebase, \n4 for init service in monorepo, \n5 for add module(s) service in monorepo, \n"+
		"6 for run multiple service in monorepo")
	flag.StringVar(&flagParam.serviceNameFlag, "servicename", "", "[project generator] define service name")
	flag.StringVar(&flagParam.packagePrefixFlag, "packageprefix", "", "[project generator] define package prefix")
	flag.BoolVar(&flagParam.withGoModFlag, "withgomod", true, "[project generator] generate go.mod or not")
	flag.StringVar(&flagParam.protoOutputPkgFlag, "protooutputpkg", "", "[project generator] define generated proto output target (if using grpc), with prefix is your go.mod")
	flag.StringVar(&flagParam.outputFlag, "output", "", "[project generator] directory to write project to (default is service name)")
	flag.StringVar(&flagParam.libraryNameFlag, "libraryname", "pkg.agungdp.dev/candi", "[project generator] define library name")

	flag.BoolVar(&flagParam.run, "run", false, "[service runner] run selected service or all service in monorepo")
	flag.StringVar(&flagParam.service, "service", "", `[service runner] depend to "-run" flag, run specific services (if multiple services, separate by comma)`)
	flag.Parse()

	switch {
	case flagParam.run:
		serviceRunner(flagParam.service)

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

		tpl = template.New(flagParam.libraryNameFlag)

		switch scope {
		case initMonorepo: // 3
			monorepoGenerator(flagParam)
		case initMonorepoService: // 4
			flagParam.packagePrefixFlag = "monorepo/services"
			flagParam.withGoModFlag = false
			flagParam.protoOutputPkgFlag = "monorepo/sdk"
			flagParam.outputFlag = "services/"
			flagParam.scopeFlag = "4"
			headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
			projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
		case addModuleMonorepoService: // 5
			flagParam.packagePrefixFlag = "monorepo/services"
			flagParam.withGoModFlag = false
			flagParam.protoOutputPkgFlag = "monorepo/sdk"
			flagParam.outputFlag = "services/"
			flagParam.scopeFlag = "5"
			headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
			projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
		case runServiceMonorepo: // 6
			serviceRunner(flagParam.service)

		default:
			headerConfig, srvConfig, modConfigs, baseConfig := parseInput(&flagParam)
			projectGenerator(flagParam, scope, headerConfig, srvConfig, modConfigs, baseConfig)
		}
	}
}
