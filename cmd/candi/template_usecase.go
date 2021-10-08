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

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/model"
	"{{$.PackagePrefix}}/pkg/shared/sharedmodel"

	"{{.LibraryName}}/candishared"
)

// {{clean (upper .ModuleName)}}Usecase abstraction
type {{clean (upper .ModuleName)}}Usecase interface {
	GetAll{{clean (upper .ModuleName)}}(ctx context.Context, filter *model.Filter{{clean (upper .ModuleName)}}) (data []sharedmodel.{{clean (upper .ModuleName)}}, meta candishared.Meta, err error)
	GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, id string) (data sharedmodel.{{clean (upper .ModuleName)}}, err error)
	Create{{clean (upper .ModuleName)}}(ctx context.Context, data *sharedmodel.{{clean (upper .ModuleName)}}) (err error)
	Update{{clean (upper .ModuleName)}}(ctx context.Context, id string, data *sharedmodel.{{clean (upper .ModuleName)}}) (err error)
	Delete{{clean (upper .ModuleName)}}(ctx context.Context, id string) (err error)
}
`
	templateUsecaseImpl = `// {{.Header}}

package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/model"
	"{{$.PackagePrefix}}/pkg/shared/sharedmodel"
	{{ if not (or .SQLDeps .MongoDeps) }}// {{end}}"{{.PackagePrefix}}/pkg/shared/repository"
	"{{$.PackagePrefix}}/pkg/shared/usecase/common"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"{{if or .KafkaHandler .RabbitMQHandler}}
	"{{.LibraryName}}/codebase/factory/types"{{end}}
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

type {{clean .ModuleName}}UsecaseImpl struct {
	sharedUsecase common.Usecase
	cache         interfaces.Cache
	{{if .SQLDeps}}repoSQL       repository.RepoSQL{{end}}
	{{if .MongoDeps}}repoMongo     repository.RepoMongo{{end}}
	{{if not .KafkaHandler}}// {{ end }}kafkaPub      interfaces.Publisher
	{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub   interfaces.Publisher
}

// New{{clean (upper .ModuleName)}}Usecase usecase impl constructor
func New{{clean (upper .ModuleName)}}Usecase(deps dependency.Dependency) ({{clean (upper .ModuleName)}}Usecase, func(sharedUsecase common.Usecase)) {
	uc := &{{clean .ModuleName}}UsecaseImpl{
		{{if .RedisDeps}}cache: deps.GetRedisPool().Cache(),{{end}}
		{{if .SQLDeps}}repoSQL:   repository.GetSharedRepoSQL(),{{end}}
		{{if .MongoDeps}}repoMongo: repository.GetSharedRepoMongo(),{{end}}
		{{if not .KafkaHandler}}// {{ end }}kafkaPub: deps.GetBroker(types.Kafka).GetPublisher(),
		{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub: deps.GetBroker(types.RabbitMQ).GetPublisher(),
	}
	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) GetAll{{clean (upper .ModuleName)}}(ctx context.Context, filter *model.Filter{{clean (upper .ModuleName)}}) (data []sharedmodel.{{clean (upper .ModuleName)}}, meta candishared.Meta, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	filter.CalculateOffset()
	{{if or .SQLDeps .MongoDeps}}data, err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().FetchAll(ctx, filter)
	if err != nil {
		return data, meta, err
	}
	count := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Count(ctx, filter)
	meta = candishared.NewMeta(filter.Page, filter.Limit, count){{end}}

	return
}

func (uc *{{clean .ModuleName}}UsecaseImpl) GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, id string) (data sharedmodel.{{clean (upper .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}repoFilter := model.Filter{{clean (upper .ModuleName)}}{ID: id}
	data, err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Find(ctx, &repoFilter){{end}}
	return
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Create{{clean (upper .ModuleName)}}(ctx context.Context, data *sharedmodel.{{clean (upper .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	return{{if or .SQLDeps .MongoDeps}} uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Save(ctx, data){{end}}
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Update{{clean (upper .ModuleName)}}(ctx context.Context, id string, data *sharedmodel.{{clean (upper .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}repoFilter := model.Filter{{clean (upper .ModuleName)}}{ID: id}
	existing, err := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Find(ctx, &repoFilter)
	if err != nil {
		return err
	}
	data.ID = existing.ID
	data.CreatedAt = existing.CreatedAt
	err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Save(ctx, data){{end}}
	return
}

func (uc *{{clean .ModuleName}}UsecaseImpl) Delete{{clean (upper .ModuleName)}}(ctx context.Context, id string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}Usecase:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	return {{if or .SQLDeps .MongoDeps}}uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{clean (upper .ModuleName)}}Repo().Delete(ctx, id){{end}}
}
`

	templateUsecaseTest = `// {{.Header}}

package usecase

import (
	"context"
	"errors"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/model"
	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	"{{$.PackagePrefix}}/pkg/shared/sharedmodel"
	"testing"

	mockdeps "{{.LibraryName}}/mocks/codebase/factory/dependency"
	mockinterfaces "{{.LibraryName}}/mocks/codebase/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew{{clean (upper .ModuleName)}}Usecase(t *testing.T) {
{{if not (or .KafkaHandler .RabbitMQHandler)}}/*{{end}}
	mockPublisher := &mockinterfaces.Publisher{}
	mockBroker := &mockinterfaces.Broker{}
	mockBroker.On("GetPublisher").Return(mockPublisher)
{{if not (or .KafkaHandler .RabbitMQHandler)}}*/{{end}}
	mockCache := &mockinterfaces.Cache{}
	mockRedisPool := &mockinterfaces.RedisPool{}
	mockRedisPool.On("Cache").Return(mockCache)

	mockDeps := &mockdeps.Dependency{}
	mockDeps.On("GetRedisPool").Return(mockRedisPool)
	{{if not (or .KafkaHandler .RabbitMQHandler)}}// {{end}}mockDeps.On("GetBroker", mock.Anything).Return(mockBroker)

	uc, setFunc := New{{clean (upper .ModuleName)}}Usecase(mockDeps)
	setFunc(nil)
	assert.NotNil(t, uc)
}

func Test_{{clean .ModuleName}}UsecaseImpl_GetAll{{clean (upper .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]sharedmodel.{{clean (upper .ModuleName)}}{}, nil)
		{{clean .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		_, _, err := uc.GetAll{{clean (upper .ModuleName)}}(context.Background(), &model.Filter{{clean (upper .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]sharedmodel.{{clean (upper .ModuleName)}}{}, errors.New("Error"))
		{{clean .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		_, _, err := uc.GetAll{{clean (upper .ModuleName)}}(context.Background(), &model.Filter{{clean (upper .ModuleName)}}{})
		assert.Error(t, err)
	})
}

func Test_{{clean .ModuleName}}UsecaseImpl_GetDetail{{clean (upper .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		responseData := sharedmodel.{{clean (upper .ModuleName)}}{}

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(responseData, nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		result, err := uc.GetDetail{{clean (upper .ModuleName)}}(context.Background(), "id")
		assert.NoError(t, err)
		assert.Equal(t, responseData, result)
	})
}

func Test_{{clean .ModuleName}}UsecaseImpl_Create{{clean (upper .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Create{{clean (upper .ModuleName)}}(context.Background(), &sharedmodel.{{clean (upper .ModuleName)}}{})
		assert.NoError(t, err)
	})
}

func Test_{{clean .ModuleName}}UsecaseImpl_Update{{clean (upper .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(sharedmodel.{{clean (upper .ModuleName)}}{}, nil)
		{{clean .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Update{{clean (upper .ModuleName)}}(context.Background(), "id", &sharedmodel.{{clean (upper .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(sharedmodel.{{clean (upper .ModuleName)}}{}, errors.New("Error"))
		{{clean .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Update{{clean (upper .ModuleName)}}(context.Background(), "id", &sharedmodel.{{clean (upper .ModuleName)}}{})
		assert.Error(t, err)
	})
}

func Test_{{clean .ModuleName}}UsecaseImpl_Delete{{clean (upper .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{clean .ModuleName}}Repo := &mockrepo.{{clean (upper .ModuleName)}}Repository{}
		{{clean .ModuleName}}Repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{clean (upper .ModuleName)}}Repo").Return({{clean .ModuleName}}Repo)

		uc := {{clean .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Delete{{clean (upper .ModuleName)}}(context.Background(), "id")
		assert.NoError(t, err)
	})
}
`
)
