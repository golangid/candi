package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const packageName = "agungdwiprasetyo.com/backend-microservices"

type param struct {
	PackageName string
	ServiceName string
	Modules     []string
}

// FileStructure model
type FileStructure struct {
	TargetDir    string
	IsDir        bool
	FromTemplate bool
	DataSource   interface{}
	Source       string
	FileName     string
	Skip         bool
	Childs       []FileStructure
}

var (
	tpl *template.Template
)

func main() {

	var scope string
	var serviceName string
	var modulesFlag string

	flag.StringVar(&scope, "scope", "initservice", "set scope")
	flag.StringVar(&serviceName, "servicename", "", "set service name")
	flag.StringVar(&modulesFlag, "modules", "", "set all modules from service")

	flag.Usage = func() {
		fmt.Println("-scope | --scope => set scope (initservice or addmodule), example: --scope initservice")
		fmt.Println("-servicename | --servicename => set service name, example: --servicename auth-service")
		fmt.Println("-modules | --modules => set modules name, example: --modules user,auth")
	}

	flag.Parse()

	var data param
	data.PackageName = packageName
	data.ServiceName = serviceName

	tpl = template.New(packageName)

	modules := strings.Split(modulesFlag, ",")
	if modulesFlag == "" {
		modules = []string{"module"} // default module name
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i] < modules[j]
	})

	apiStructure := FileStructure{
		TargetDir: "api/{{.ServiceName}}/", IsDir: true, DataSource: data,
	}

	cmdStructure := FileStructure{
		TargetDir: "cmd/{{.ServiceName}}/", IsDir: true, DataSource: data,
		Childs: []FileStructure{
			{FromTemplate: true, DataSource: data, Source: cmdMainTemplate, FileName: "main.go"},
			{FromTemplate: true, DataSource: data, Source: envTemplate, FileName: ".env"},
			{FromTemplate: true, DataSource: data, Source: envTemplate, FileName: ".env.sample"},
		},
	}

	serviceStructure := FileStructure{
		TargetDir: "internal/{{.ServiceName}}/", IsDir: true, DataSource: data,
	}

	apiProtoStructure := FileStructure{
		TargetDir: "proto/", IsDir: true,
	}
	apiGraphQLStructure := FileStructure{
		TargetDir: "graphql/", IsDir: true,
	}

	if scope == "addmodule" {
		files, err := ioutil.ReadDir("internal/" + serviceName + "/modules")
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if f.IsDir() {
				data.Modules = append(data.Modules, f.Name())
			}
		}
	}

	var moduleStructure = FileStructure{
		TargetDir: "modules/", IsDir: true, DataSource: data,
	}
	for _, moduleName := range modules {
		moduleName = strings.TrimSpace(moduleName)
		data.Modules = append(data.Modules, moduleName)
		dataSource := map[string]string{"PackageName": packageName, "ServiceName": serviceName, "module": moduleName}

		cleanArchModuleDir := []FileStructure{
			{
				TargetDir: "delivery/", IsDir: true,
				Childs: []FileStructure{
					{TargetDir: "graphqlhandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, DataSource: dataSource, Source: deliveryGraphqlTemplate, FileName: "graphqlhandler.go"},
					}},
					{TargetDir: "grpchandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, DataSource: dataSource, Source: deliveryGRPCTemplate, FileName: "grpchandler.go"},
					}},
					{TargetDir: "resthandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, DataSource: dataSource, Source: deliveryRestTemplate, FileName: "resthandler.go"},
					}},
					{TargetDir: "workerhandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, DataSource: dataSource, Source: deliveryKafkaTemplate, FileName: "kafkahandler.go"},
					}},
				},
			},
			{
				TargetDir: "domain/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, FileName: "domain.go"},
				},
			},
			{
				TargetDir: "repository/", IsDir: true,
				Childs: []FileStructure{
					{TargetDir: "interfaces/", IsDir: true},
					{FromTemplate: true, FileName: "repository.go"},
				},
			},
			{
				TargetDir: "usecase/", IsDir: true,
				Childs: []FileStructure{
					{FromTemplate: true, FileName: "usecase.go"},
					{FromTemplate: true, FileName: "usecase_impl.go"},
				},
			},
		}

		moduleStructure.Childs = append(moduleStructure.Childs, []FileStructure{
			{
				TargetDir: moduleName + "/", IsDir: true,
				Childs: append(cleanArchModuleDir,
					FileStructure{
						FromTemplate: true, DataSource: dataSource, Source: moduleMainTemplate, FileName: "module.go",
					},
				),
			},
		}...)

		apiProtoStructure.Childs = append(apiProtoStructure.Childs, FileStructure{
			TargetDir: moduleName, IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, DataSource: dataSource, Source: defaultGRPCProto, FileName: moduleName + ".proto"},
			},
		})
		apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, FileStructure{
			FromTemplate: true, DataSource: dataSource, Source: defaultGraphqlSchema, FileName: moduleName + ".graphql",
		})
	}
	serviceStructure.Childs = append(serviceStructure.Childs, moduleStructure)
	serviceStructure.Childs = append(serviceStructure.Childs, FileStructure{
		FromTemplate: true, DataSource: data, Source: serviceMainTemplate, FileName: "service.go",
	})

	apiGraphQLStructure.Childs = append(apiGraphQLStructure.Childs, FileStructure{
		FromTemplate: true, DataSource: data, Source: defaultGraphqlRootSchema, FileName: "_schema.graphql",
	})
	apiStructure.Childs = []FileStructure{
		apiGraphQLStructure,
		{
			TargetDir: "jsonschema/", IsDir: true,
			Childs: []FileStructure{
				// {FromTemplate: true, DataSource: dataSource, FileName: "_schema.json"},
			},
		},
		apiProtoStructure,
	}

	var baseDirectoryFile FileStructure
	switch scope {
	case "initservice":
		baseDirectoryFile.Childs = []FileStructure{
			apiStructure, cmdStructure, serviceStructure,
		}

	case "addmodule":
		moduleStructure.Skip = true
		serviceStructure.Skip = true
		serviceStructure.Childs = []FileStructure{
			moduleStructure,
			{FromTemplate: true, DataSource: data, Source: serviceMainTemplate, FileName: "service.go"},
		}

		apiStructure.Skip = true
		apiProtoStructure.Skip, apiGraphQLStructure.Skip = true, true
		apiStructure.Childs = []FileStructure{
			apiProtoStructure, apiGraphQLStructure,
		}

		baseDirectoryFile.Childs = []FileStructure{apiStructure, serviceStructure}
		baseDirectoryFile.Skip = true

	default:
		panic("invalid scope parameter")
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
	replacer := strings.NewReplacer("-", "", "*", "", "/", "", ":", "")
	return template.FuncMap{

		"clean": func(v string) string {
			return replacer.Replace(v)
		},

		"upper": func(str string) string {
			return strings.Title(str)
		},
	}
}
