package main

const (
	templateUsecaseUOW = `// {{.Header}}

package usecase

import (
	"sync"

{{- range $module := .Modules}}
	{{clean $module.ModuleName}}usecase "{{$.PackagePrefix}}/internal/modules/{{cleanPathModule $module.ModuleName}}/usecase"
{{- end }}

	"{{.LibraryName}}/codebase/factory/dependency"
)

type (
	// Usecase unit of work for all usecase in modules
	Usecase interface {
	{{- range $module := .Modules}}
		{{clean (upper $module.ModuleName)}}() {{clean $module.ModuleName}}usecase.{{clean (upper $module.ModuleName)}}Usecase
	{{- end }}
	}

	usecaseUow struct {
	{{- range $module := .Modules}}
		{{clean $module.ModuleName}} {{clean $module.ModuleName}}usecase.{{clean (upper $module.ModuleName)}}Usecase
	{{- end }}
	}
)

var usecaseInst *usecaseUow
var once sync.Once

// SetSharedUsecase set singleton usecase unit of work instance
func SetSharedUsecase(deps dependency.Dependency) {
	once.Do(func() {
		usecaseInst = &usecaseUow{
		{{- range $module := .Modules}}
			{{clean $module.ModuleName}}: {{clean $module.ModuleName}}usecase.New{{clean (upper $module.ModuleName)}}Usecase(deps),
		{{- end }}
		}
	})
}

// GetSharedUsecase get usecase unit of work instance
func GetSharedUsecase() Usecase {
	return usecaseInst
}

{{- range $module := .Modules}}
func (uc *usecaseUow) {{clean (upper $module.ModuleName)}}() {{clean $module.ModuleName}}usecase.{{clean (upper $module.ModuleName)}}Usecase {
	return uc.{{clean $module.ModuleName}}
}
{{- end }}
`

	templateUsecaseAbstraction = `// {{.Header}}

package usecase

import (
	"context"
)

// {{clean (upper .ModuleName)}}Usecase abstraction
type {{clean (upper .ModuleName)}}Usecase interface {
	// add method
	Hello(ctx context.Context) string
}
`
	templateUsecaseImpl = `// {{.Header}}

package usecase

import (
	"context"

	{{ if not (or .SQLDeps .MongoDeps) }}// {{end}}"{{.PackagePrefix}}/pkg/shared/repository"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

type {{clean .ModuleName}}UsecaseImpl struct {
	cache interfaces.Cache
	{{if .SQLDeps}}repoSQL   *repository.RepoSQL{{end}}
	{{if .MongoDeps}}repoMongo *repository.RepoMongo{{end}}
}

// New{{clean (upper .ModuleName)}}Usecase usecase impl constructor
func New{{clean (upper .ModuleName)}}Usecase(deps dependency.Dependency) {{clean (upper .ModuleName)}}Usecase {
	return &{{clean .ModuleName}}UsecaseImpl{
		{{if .RedisDeps}}cache: deps.GetRedisPool().Cache(),{{end}}
		{{if .SQLDeps}}repoSQL:   repository.GetSharedRepoSQL(),{{end}}
		{{if .MongoDeps}}repoMongo: repository.GetSharedRepoMongo(),{{end}}
	}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Hello(ctx context.Context) (msg string) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}Usecase:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	{{if .SQLDeps}}msg, _ = uc.repoSQL.{{clean (upper .ModuleName)}}Repo.FindHello(ctx){{end}}
	{{if .MongoDeps}}msg, _ = uc.repoMongo.{{clean (upper .ModuleName)}}Repo.FindHello(ctx){{end}}
	return
}	
`
)
