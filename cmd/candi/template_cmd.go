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

	"{{$.PackagePrefix}}/cmd/migration/migrations"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/config/database"
	"{{.LibraryName}}/config/env"

	"github.com/pressly/goose/v3"

	{{if eq .SQLDriver "postgres"}}_ "github.com/lib/pq"{{else if eq .SQLDriver "mysql"}}_ "github.com/go-sql-driver/mysql"{{end}}
	"gorm.io/driver/{{.SQLDriver}}"
	"gorm.io/gorm"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
)

func main() {
	env.Load("{{.ServiceName}}")
	sqlDeps := database.InitSQLDatabase()

	flags.Parse(os.Args[1:])
	args := flags.Args()
	arguments := []string{}
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	dir := os.Getenv("WORKDIR") + "cmd/migration/migrations"
	switch args[0] {
	case "create":
		migrationType := "sql"
		if len(args) > 2 && args[2] == "init_table" {
			migrationType = "go"
		}
		if err := goose.Create(sqlDeps.WriteDB(), dir, args[1], migrationType); err != nil {
			log.Fatalf("goose %v: %v", args[1], err)
		}

	default:

		if err := goose.Run(args[0], sqlDeps.WriteDB(), dir, arguments...); err != nil {
			log.Fatalf("goose %v: %v", args[0], err)
		}

		if migrateTables := migrations.GetMigrateTables(); len(migrateTables) > 0 {
			gormWrite, err := gorm.Open({{ .SQLDriver }}.New({{ .SQLDriver }}.Config{
				Conn: sqlDeps.WriteDB(),
			}), &gorm.Config{
				SkipDefaultTransaction:                   true,
				DisableForeignKeyConstraintWhenMigrating: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			tx := gormWrite.Begin()
			if err := gormWrite.AutoMigrate(migrateTables...); err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
			tx.Commit()
		}
	}
	log.Printf("\x1b[32;1mMigration to \"%s\" success\x1b[0m\n", candihelper.MaskingPasswordURL(env.BaseEnv().DbSQLWriteDSN))
}
`

	templateCmdMigrationInitTable = `package migrations

var (
	migrateTables []interface{}
)

// GetMigrateTables get migrate table list
func GetMigrateTables() []interface{} {
	return migrateTables
}
`

	templateCmdMigrationInitModule = `-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS {{snake .ModuleName}}s (
	"id" varchar(255) NOT NULL PRIMARY KEY,
	"field" varchar(255),
	"created_at" timestamptz(6),
	"modified_at" timestamptz(6)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS {{snake .ModuleName}}s;
-- +goose StatementEnd
`
)
