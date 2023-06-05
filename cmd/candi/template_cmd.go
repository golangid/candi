package main

const (
	cmdMainTemplate = `// {{.Header}}

package main

import (
	"fmt"
	"runtime/debug"{{if eq .SQLDriver "postgres"}}
	_ "github.com/lib/pq"{{else if eq .SQLDriver "mysql"}}
	_ "github.com/go-sql-driver/mysql"{{else if eq .SQLDriver "sqlite3"}}
	_ "github.com/mattn/go-sqlite3"{{end}}

	"{{.LibraryName}}/codebase/app"
	"{{.LibraryName}}/config"

	service "{{.PackagePrefix}}/internal"
)

func main() {
	const serviceName = "{{.ServiceName}}"

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

	{{if eq .SQLDriver "postgres"}}_ "github.com/lib/pq"{{else if eq .SQLDriver "mysql"}}_ "github.com/go-sql-driver/mysql"{{else if eq .SQLDriver "sqlite3"}}_ "github.com/mattn/go-sqlite3"{{end}}
	{{if eq .SQLDriver "sqlite3"}}"gorm.io/driver/sqlite"{{else}}"gorm.io/driver/{{.SQLDriver}}"{{end}}
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
		if len(args) > 2 {
			migrationType = args[2]
		}
		if err := goose.Create(sqlDeps.WriteDB(), dir, args[1], migrationType); err != nil {
			log.Fatalf("goose %v: %v", args[1], err)
		}

	default:
		if err := goose.SetDialect("{{ .SQLDriver }}"); err != nil {
			log.Fatal(err)
		}
		if err := goose.Run(args[0], sqlDeps.WriteDB(), dir, arguments...); err != nil {
			log.Fatalf("goose %v: %v", args[0], err)
		}

		if migrateTables := migrations.GetMigrateTables(); len(migrateTables) > 0 {
			gormWrite, err := gorm.Open({{if eq .SQLDriver "sqlite3"}}sqlite.Dialector{Conn: sqlDeps.WriteDB()}{{else}}{{ .SQLDriver }}.New({{ .SQLDriver }}.Config{
				Conn: sqlDeps.WriteDB(),
			}){{end}}, &gorm.Config{
				SkipDefaultTransaction:                   true,
				DisableForeignKeyConstraintWhenMigrating: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			tx := gormWrite.Begin()
			if err := tx.AutoMigrate(migrateTables...); err != nil {
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
{{if eq .SQLDriver "mysql"}}CREATE TABLE IF NOT EXISTS ` + "`" + `{{plural .ModuleName}}` + "`" + ` (
	` + "`" + `id` + "`" + ` SERIAL NOT NULL PRIMARY KEY,
	` + "`" + `field` + "`" + ` VARCHAR(255),
	` + "`" + `created_at` + "`" + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	` + "`" + `updated_at` + "`" + ` TIMESTAMP NULL
);{{else}}CREATE TABLE IF NOT EXISTS {{plural .ModuleName}} (
	"id" SERIAL NOT NULL PRIMARY KEY,
	"field" VARCHAR(255),
	"created_at" TIMESTAMPTZ(6),
	"updated_at" TIMESTAMPTZ(6)
);{{end}}
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS {{plural .ModuleName}};
-- +goose StatementEnd
`
)
