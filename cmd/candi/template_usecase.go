package main

const (
	templateUsecaseUOW = `// {{.Header}}

package usecase

import (
	"sync"

	// @candi:usecaseImport
	"{{$.PackagePrefix}}/pkg/shared/usecase/common"

	"{{.LibraryName}}/codebase/factory/dependency"
)

type (
	// Usecase unit of work for all usecase in modules
	Usecase interface {
		// @candi:usecaseMethod
	}

	usecaseUow struct {
		// @candi:usecaseField
	}
)

var usecaseInst *usecaseUow
var once sync.Once

// SetSharedUsecase set singleton usecase unit of work instance
func SetSharedUsecase(deps dependency.Dependency) {
	once.Do(func() {
		usecaseInst = new(usecaseUow)
		var setSharedUsecaseFuncs []func(common.Usecase)
		var setSharedUsecaseFunc func(common.Usecase)

		// @candi:usecaseCommon

		sharedUsecase := common.SetCommonUsecase(usecaseInst)
		for _, setFunc := range setSharedUsecaseFuncs {
			setFunc(sharedUsecase)
		}
	})
}

// GetSharedUsecase get usecase unit of work instance
func GetSharedUsecase() Usecase {
	return usecaseInst
}

// @candi:usecaseImplementation
`

	templateUsecaseCommon = `// {{.Header}}
	
package common

var commonUC Usecase

// Usecase common abstraction for bridging shared method inter usecase in module
type Usecase interface {
	// shared method from another usecase
}

// SetCommonUsecase constructor
func SetCommonUsecase(uc Usecase) Usecase {
	commonUC = uc
	return commonUC
}

// GetCommonUsecase get common usecase
func GetCommonUsecase() Usecase {
	return commonUC
}
`

	templateUsecaseAbstraction = `// {{.Header}}

package usecase

import (
	"context"

	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
)

// {{clean (upper .ModuleName)}}Usecase abstraction
type {{clean (upper .ModuleName)}}Usecase interface {
	GetAll{{clean (upper .ModuleName)}}(ctx context.Context, filter candishared.Filter) (data []shareddomain.{{clean (upper .ModuleName)}}, meta candishared.Meta, err error)
	GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, id string) (data shareddomain.{{clean (upper .ModuleName)}}, err error)
	Create{{clean (upper .ModuleName)}}(ctx context.Context, data *shareddomain.{{clean (upper .ModuleName)}}) (err error)
	Update{{clean (upper .ModuleName)}}(ctx context.Context, id string, data *shareddomain.{{clean (upper .ModuleName)}}) (err error)
	Delete{{clean (upper .ModuleName)}}(ctx context.Context, id string) (err error)
}
`
	templateUsecaseImpl = `// {{.Header}}

package usecase

import (
	"context"

	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	{{ if not (or .SQLDeps .MongoDeps) }}// {{end}}"{{.PackagePrefix}}/pkg/shared/repository"
	"{{$.PackagePrefix}}/pkg/shared/usecase/common"

	"github.com/google/uuid"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

type {{clean .ModuleName}}UsecaseImpl struct {
	sharedUsecase common.Usecase
	cache         interfaces.Cache
	{{if .SQLDeps}}repoSQL   repository.RepoSQL{{end}}
	{{if .MongoDeps}}repoMongo repository.RepoMongo{{end}}
}

// New{{clean (upper .ModuleName)}}Usecase usecase impl constructor
func New{{clean (upper .ModuleName)}}Usecase(deps dependency.Dependency) ({{clean (upper .ModuleName)}}Usecase, func(sharedUsecase common.Usecase)) {
	uc := &{{clean .ModuleName}}UsecaseImpl{
		{{if .RedisDeps}}cache: deps.GetRedisPool().Cache(),{{end}}
		{{if .SQLDeps}}repoSQL:   repository.GetSharedRepoSQL(),{{end}}
		{{if .MongoDeps}}repoMongo: repository.GetSharedRepoMongo(),{{end}}
	}
	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) GetAll{{clean (upper .ModuleName)}}(ctx context.Context, filter candishared.Filter) (data []shareddomain.{{clean (upper .ModuleName)}}, meta candishared.Meta, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}data, err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().FetchAll(ctx, &filter)
	if err != nil {
		return data, meta, err
	}
	count := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Count(ctx, &filter)
	meta = candishared.NewMeta(filter.Page, filter.Limit, count){{end}}

	return
}

func (uc *{{clean .ModuleName}}UsecaseImpl) GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, id string) (data shareddomain.{{clean (upper .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}data.ID = id
	err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Find(ctx, &data){{end}}
	return
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Create{{clean (upper .ModuleName)}}(ctx context.Context, data *shareddomain.{{clean (upper .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	data.ID = uuid.NewString()
	return{{if or .SQLDeps .MongoDeps}} uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Save(ctx, data){{end}}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Update{{clean (upper .ModuleName)}}(ctx context.Context, id string, data *shareddomain.{{clean (upper .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	existing := &shareddomain.{{clean (upper .ModuleName)}}{ID: id}
	{{if or .SQLDeps .MongoDeps}}if err := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Find(ctx, existing); err != nil {
		return err
	}{{end}}
	data.ID = existing.ID
	data.CreatedAt = existing.CreatedAt
	return {{if or .SQLDeps .MongoDeps}} uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Save(ctx, data){{end}}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Delete{{clean (upper .ModuleName)}}(ctx context.Context, id string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	return{{if or .SQLDeps .MongoDeps}} uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Delete(ctx, id){{end}}
}
`
)
