package main

const (
	cmdMainTemplate = `// {{.Header}}

package main

import (
	"fmt"
	"runtime/debug"{{if eq .SQLDriver "postgres"}}
	_ "github.com/lib/pq"{{else if eq .SQLDriver "mysql"}}
	_ "github.com/go-sql-driver/mysql"{{end}}

	"{{.LibraryName}}/codebase/app"
	"{{.LibraryName}}/config"

	service "{{.PackagePrefix}}/internal"
)

const serviceName = "{{.ServiceName}}"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mFailed to start %s service: %v\x1b[0m\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(serviceName)
	defer cfg.Exit()

	app.New(service.NewService(cfg)).Run()
}
`

	templateCmdMigration = `// {{.Header}}

package main

import (
	"flag"
	"log"
	"os"

	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/config/database"
	"{{.LibraryName}}/config/env"

	{{if eq .SQLDriver "postgres"}}_ "github.com/lib/pq"{{else if eq .SQLDriver "mysql"}}_ "github.com/go-sql-driver/mysql"{{end}}
	"gorm.io/driver/{{.SQLDriver}}"
	"gorm.io/gorm"
)

func main() {
	var databaseConnection string
	flag.StringVar(&databaseConnection, "dbconn", "", "Database connection target")
	flag.Parse()

	if databaseConnection == "" {
		flag.Usage()
		os.Exit(1)
	}

	env.SetEnv(env.Env{DbSQLWriteDSN: databaseConnection})
	sqlDeps := database.InitSQLDatabase()
	gormWrite, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDeps.WriteDB(),
	}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	gormWrite.AutoMigrate({{- range $module := .Modules}}
		&shareddomain.{{clean (upper $module.ModuleName)}}{},{{- end}}
	)
	log.Printf("\x1b[32;1mMigration to \"%s\" suceess\x1b[0m\n", candihelper.MaskingPasswordURL(databaseConnection))
}
`
)
