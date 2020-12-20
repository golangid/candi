package main

const (
	templateRepository = `// {{.Header}} DO NOT EDIT.

package repository

import (
	"sync"

	"{{.PackageName}}/codebase/factory/dependency"
)

var (
	once sync.Once
)

// SetSharedRepository set the global singleton "RepoSQL" and "RepoMongo" implementation
func SetSharedRepository(deps dependency.Dependency) {
	once.Do(func() {
		{{if not .SQLDeps}}// {{end}}setSharedRepoSQL(deps.GetSQLDatabase().ReadDB(), deps.GetSQLDatabase().WriteDB())
		{{if not .MongoDeps}}// {{end}}setSharedRepoMongo(deps.GetMongoDatabase().ReadDB(), deps.GetMongoDatabase().WriteDB())
	})
}
`

	templateRepositoryUOWSQL = `// {{.Header}}

package repository

import (
	"context"
	"database/sql"
	"fmt"

{{- range $module := .Modules}}
	{{if $module.SQLDeps}}{{clean $module.ModuleName}}repo "{{$.GoModName}}/internal/modules/{{cleanPathModule $module.ModuleName}}/repository"{{end}}
{{- end }}

	"{{.PackageName}}/tracer"
)

// RepoSQL uow
type RepoSQL struct {
	readDB, writeDB *sql.DB
	tx              *sql.Tx

	// register all repository from modules
{{- range $module := .Modules}}
	{{if $module.SQLDeps}}{{clean (upper $module.ModuleName)}}Repo {{clean $module.ModuleName}}repo.{{clean (upper $module.ModuleName)}}Repository{{end}}
{{- end }}
}

var (
	globalRepoSQL = new(RepoSQL)
)

// setSharedRepoSQL set the global singleton "RepoSQL" implementation
func setSharedRepoSQL(readDB, writeDB *sql.DB) {
	globalRepoSQL = NewRepositorySQL(readDB, writeDB, nil)
}

// GetSharedRepoSQL returns the global singleton "RepoSQL" implementation
func GetSharedRepoSQL() *RepoSQL {
	return globalRepoSQL
}

// NewRepositorySQL constructor
func NewRepositorySQL(readDB, writeDB *sql.DB, tx *sql.Tx) *RepoSQL {

	return &RepoSQL{
		readDB: readDB, writeDB: writeDB, tx: tx,

{{- range $module := .Modules}}
		{{if $module.SQLDeps}}{{clean (upper $module.ModuleName)}}Repo: {{clean $module.ModuleName}}repo.New{{clean (upper $module.ModuleName)}}RepoSQL(readDB, writeDB, tx),{{end}}
{{- end }}
	}
}

// WithTransaction run transaction for each repository with context, include handle canceled or timeout context
func (r *RepoSQL) WithTransaction(ctx context.Context, txFunc func(ctx context.Context, repo *RepoSQL) error) (err error) {
	trace := tracer.StartTrace(ctx, "RepoSQL-Transaction")
	defer trace.Finish()
	ctx = trace.Context()

	tx, errInit := r.writeDB.Begin()
	if errInit != nil {
		return errInit
	}

	// reinit new repository in different memory address with tx value
	manager := NewRepositorySQL(r.readDB, r.writeDB, tx)
	defer manager.free()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		if err != nil {
			tx.Rollback()
			trace.SetError(err)
		} else {
			tx.Commit()
		}
	}()

	errChan := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
			close(errChan)
		}()

		if err := txFunc(ctx, manager); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Canceled or timeout: %v", ctx.Err())
	case e := <-errChan:
		return e
	}
}

func (r *RepoSQL) free() {
	// make nil all repository
{{- range $module := .Modules}}
	{{if $module.SQLDeps}}r.{{clean (upper $module.ModuleName)}}Repo = nil{{end}}
{{- end }}
}	
`

	templateRepositoryUOWMongo = `// {{.Header}} DO NOT EDIT.

package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

{{- range $module := .Modules}}
	{{if $module.MongoDeps}}{{clean $module.ModuleName}}repo "{{$.GoModName}}/internal/modules/{{cleanPathModule $module.ModuleName}}/repository"{{end}}
{{- end }}
)

// RepoMongo uow
type RepoMongo struct {
	readDB, writeDB *mongo.Database

	// register all repository from modules
{{- range $module := .Modules}}
	{{if $module.MongoDeps}}{{clean (upper $module.ModuleName)}}Repo {{clean $module.ModuleName}}repo.{{clean (upper $module.ModuleName)}}Repository{{end}}
{{- end }}
}

var globalRepoMongo = new(RepoMongo)

// setSharedRepoMongo set the global singleton "RepoMongo" implementation
func setSharedRepoMongo(readDB, writeDB *mongo.Database) {
	globalRepoMongo = &RepoMongo{
		readDB: readDB, writeDB: writeDB,

{{- range $module := .Modules}}
		{{if $module.MongoDeps}}{{clean (upper $module.ModuleName)}}Repo: {{clean $module.ModuleName}}repo.New{{clean (upper $module.ModuleName)}}RepoMongo(readDB, writeDB),{{end}}
{{- end }}
	}
}

// GetSharedRepoMongo returns the global singleton "RepoMongo" implementation
func GetSharedRepoMongo() *RepoMongo {
	return globalRepoMongo
}
`

	templateRepositoryAbstraction = `// {{.Header}}

package repository

import (
	"context"
)

// {{clean (upper .ModuleName)}}Repository abstract interface
type {{clean (upper .ModuleName)}}Repository interface {
	// add method
	FindHello(ctx context.Context) (string, error)
}
`

	templateRepositoryMongoImpl = `// {{.Header}}

package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"{{.PackageName}}/tracer"
)

type {{clean .ModuleName}}RepoMongo struct {
	readDB, writeDB *mongo.Database
}

// New{{clean (upper .ModuleName)}}RepoMongo mongo repo constructor
func New{{clean (upper .ModuleName)}}RepoMongo(readDB, writeDB *mongo.Database) {{clean (upper .ModuleName)}}Repository {
	return &{{clean .ModuleName}}RepoMongo{
		readDB, writeDB,
	}
}

func (r *{{clean .ModuleName}}RepoMongo) FindHello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}RepoMongo:FindHello")
	defer trace.Finish()

	return "Hello from repo mongo layer", nil
}
`

	templateRepositorySQLImpl = `// {{.Header}}

package repository

import (
	"context"
	"database/sql"

	"{{.PackageName}}/tracer"
)

type {{clean .ModuleName}}RepoSQL struct {
	readDB, writeDB *sql.DB
	tx              *sql.Tx
}

// New{{clean (upper .ModuleName)}}RepoSQL mongo repo constructor
func New{{clean (upper .ModuleName)}}RepoSQL(readDB, writeDB *sql.DB, tx *sql.Tx) {{clean (upper .ModuleName)}}Repository {
	return &{{clean .ModuleName}}RepoSQL{
		readDB, writeDB, tx,
	}
}

func (r *{{clean .ModuleName}}RepoSQL) FindHello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}RepoSQL:FindHello")
	defer trace.Finish()

	return "Hello from repo sql layer", nil
}`
)
