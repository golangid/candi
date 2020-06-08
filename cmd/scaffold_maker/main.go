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
	Source       string
	FileName     string
	Childs       []FileStructure
}

var (
	data param

	baseDirectoryFileList = []FileStructure{
		FileStructure{
			TargetDir: "api/{{.ServiceName}}/", IsDir: true,
			Childs: []FileStructure{
				{TargetDir: "graphql/", IsDir: true},
				{TargetDir: "jsonschema/", IsDir: true},
				{TargetDir: "proto/", IsDir: true},
			},
		},
		FileStructure{
			TargetDir: "cmd/{{.ServiceName}}/", IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, Source: cmdMainTemplate, FileName: "main.go"},
				{FromTemplate: false, Source: "# Additional service environment", FileName: ".env.extend"},
			},
		},
	}

	serviceStructure = FileStructure{
		TargetDir: "internal/services/{{.ServiceName}}/", IsDir: true,
		Childs: []FileStructure{
			{FromTemplate: true, Source: serviceMainTemplate, FileName: "service.go"},
		},
	}

	cleanArchModuleDir = []FileStructure{
		FileStructure{
			TargetDir: "delivery/", IsDir: true,
			Childs: []FileStructure{
				{TargetDir: "graphqlhandler/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, FileName: "graphqlhandler.go"},
				}},
				{TargetDir: "grpchandler/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, FileName: "grpchandler.go"},
				}},
				{TargetDir: "resthandler/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, FileName: "resthandler.go"},
				}},
				{TargetDir: "subscriberhandler/", IsDir: true, Childs: []FileStructure{
					{FromTemplate: true, FileName: "subscriberhandler.go"},
				}},
			},
		},
		FileStructure{
			TargetDir: "domain/", IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, FileName: "domain.go"},
			},
		},
		FileStructure{
			TargetDir: "repository/", IsDir: true,
			Childs: []FileStructure{
				{TargetDir: "interfaces/", IsDir: true},
				{FromTemplate: true, FileName: "repository.go"},
			},
		},
		FileStructure{
			TargetDir: "usecase/", IsDir: true,
			Childs: []FileStructure{
				{FromTemplate: true, FileName: "usecase.go"},
				{FromTemplate: true, FileName: "usecase_impl.go"},
			},
		},
	}

	tpl *template.Template
)

func main() {

	var serviceName string
	var modulesFlag string

	flag.StringVar(&serviceName, "servicename", "", "set service name")
	flag.StringVar(&modulesFlag, "modules", "", "set all modules from service")

	flag.Usage = func() {
		fmt.Println("-servicename | --servicename => set service name, example: --servicename auth-service")
		fmt.Println("-modules | --modules => set service name, example: --modules user,auth")
	}

	flag.Parse()

	tpl = template.New(packageName)

	modules := strings.Split(modulesFlag, ",")
	if modulesFlag == "" {
		modules = []string{"module"} // default module name
	}

	var moduleStructure = FileStructure{
		TargetDir: "modules/", IsDir: true,
	}
	for _, module := range modules {
		data.Modules = append(data.Modules, module)
		buff := loadTemplate(moduleMainTemplate, map[string]string{"PackageName": packageName, "module": module})
		moduleStructure.Childs = append(moduleStructure.Childs, []FileStructure{
			{TargetDir: module + "/", IsDir: true, Childs: append(cleanArchModuleDir, FileStructure{
				FromTemplate: false, Source: string(buff), FileName: "module.go",
			})},
		}...)
	}

	serviceStructure.Childs = append(serviceStructure.Childs, moduleStructure)
	baseDirectoryFileList = append(baseDirectoryFileList, serviceStructure)

	data.PackageName = packageName
	data.ServiceName = serviceName

	for _, fl := range baseDirectoryFileList {
		exec(fl, 0)
	}

}

func exec(fl FileStructure, depth int) {
	dirBuff := loadTemplate(fl.TargetDir, data)

	dirName := string(dirBuff)
	if depth == 0 {
		fmt.Println(dirName, depth)
		if _, err := os.Stat(dirName); os.IsExist(err) {
			return
		}
	}

	if fl.IsDir {
		fmt.Println("mkdir", dirName)
		if err := os.Mkdir(dirName, 0700); err != nil {
			fmt.Println("mkdir err:", err)
			panic(err)
		}
	}

	if fl.FileName != "" {
		var buff []byte
		if fl.FromTemplate {
			if fl.Source != "" {
				buff = loadTemplate(fl.Source, data)
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
	replacer := strings.NewReplacer("-", "", "*", "", "/", "", "*", "")
	return template.FuncMap{

		"clean": func(v interface{}) string {
			return replacer.Replace(fmt.Sprint(v))
		},
	}
}
