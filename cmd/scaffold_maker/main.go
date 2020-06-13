package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	Childs       []FileStructure
}

var (
	tpl *template.Template
)

func main() {

	var serviceName string
	var modulesFlag string

	flag.StringVar(&serviceName, "servicename", "", "set service name")
	flag.StringVar(&modulesFlag, "modules", "", "set all modules from service")

	flag.Usage = func() {
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

	apiStructure := FileStructure{
		TargetDir: "api/{{.ServiceName}}/", IsDir: true, DataSource: data,
		Childs: []FileStructure{
			{TargetDir: "graphql/", IsDir: true},
			{TargetDir: "jsonschema/", IsDir: true},
			{TargetDir: "proto/", IsDir: true},
		},
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
		TargetDir: "internal/services/{{.ServiceName}}/", IsDir: true, DataSource: data,
	}

	var moduleStructure = FileStructure{
		TargetDir: "modules/", IsDir: true, DataSource: data,
	}
	for _, moduleName := range modules {
		data.Modules = append(data.Modules, moduleName)
		dataSource := map[string]string{"PackageName": packageName, "ServiceName": serviceName, "module": moduleName}

		cleanArchModuleDir := []FileStructure{
			{
				TargetDir: "delivery/", IsDir: true,
				Childs: []FileStructure{
					{TargetDir: "graphqlhandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, FileName: "graphqlhandler.go"},
					}},
					{TargetDir: "grpchandler/", IsDir: true, Childs: []FileStructure{
						{FromTemplate: true, FileName: "grpchandler.go"},
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
	}
	serviceStructure.Childs = append(serviceStructure.Childs, moduleStructure)
	serviceStructure.Childs = append(serviceStructure.Childs, FileStructure{
		FromTemplate: true, DataSource: data, Source: serviceMainTemplate, FileName: "service.go"},
	)

	baseDirectoryFileList := []FileStructure{
		apiStructure, cmdStructure, serviceStructure,
	}

	for _, fl := range baseDirectoryFileList {
		exec(fl, 0)
	}

}

func exec(fl FileStructure, depth int) {
	dirBuff := loadTemplate(fl.TargetDir, fl.DataSource)

	dirName := string(dirBuff)
	if depth == 0 {
		fmt.Println(dirName, depth)
		if _, err := os.Stat(dirName); os.IsExist(err) {
			panic(err)
		}
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

	for _, child := range fl.Childs {
		child.TargetDir = dirName + child.TargetDir
		exec(child, depth+1)
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

		"clean": func(v interface{}) string {
			return replacer.Replace(fmt.Sprint(v))
		},
	}
}

// func parse
