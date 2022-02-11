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

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
)

// {{upper (camel .ModuleName)}}Usecase abstraction
type {{upper (camel .ModuleName)}}Usecase interface {
	GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, meta candishared.Meta, err error)
	GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id string) (data shareddomain.{{upper (camel .ModuleName)}}, err error)
	Create{{upper (camel .ModuleName)}}(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error)
	Update{{upper (camel .ModuleName)}}(ctx context.Context, id string, data *shareddomain.{{upper (camel .ModuleName)}}) (err error)
	Delete{{upper (camel .ModuleName)}}(ctx context.Context, id string) (err error)
}
`
	templateUsecaseImpl = `// {{.Header}}

package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	{{ if not (or .SQLDeps .MongoDeps .ArangoDeps) }}// {{end}}"{{.PackagePrefix}}/pkg/shared/repository"
	"{{$.PackagePrefix}}/pkg/shared/usecase/common"
	{{if and .MongoDeps (not .SQLDeps)}}
	"go.mongodb.org/mongo-driver/bson/primitive"
	{{end}}
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"{{if or .KafkaHandler .RabbitMQHandler}}
	"{{.LibraryName}}/codebase/factory/types"{{end}}
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

type {{camel .ModuleName}}UsecaseImpl struct {
	sharedUsecase common.Usecase
	cache         interfaces.Cache
	{{if .SQLDeps}}repoSQL       repository.RepoSQL{{end}}
	{{if .MongoDeps}}repoMongo     repository.RepoMongo{{end}}
	{{if .ArangoDeps}}repoArango     repository.RepoArango{{end}}
	{{if not .KafkaHandler}}// {{ end }}kafkaPub      interfaces.Publisher
	{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub   interfaces.Publisher
}

// New{{upper (camel .ModuleName)}}Usecase usecase impl constructor
func New{{upper (camel .ModuleName)}}Usecase(deps dependency.Dependency) ({{upper (camel .ModuleName)}}Usecase, func(sharedUsecase common.Usecase)) {
	uc := &{{camel .ModuleName}}UsecaseImpl{
		{{if .RedisDeps}}cache: deps.GetRedisPool().Cache(),{{end}}
		{{if .SQLDeps}}repoSQL:   repository.GetSharedRepoSQL(),{{end}}
		{{if .MongoDeps}}repoMongo: repository.GetSharedRepoMongo(),{{end}}
		{{if .ArangoDeps}}repoArango: repository.GetSharedRepoArango(),{{end}}
		{{if not .KafkaHandler}}// {{ end }}kafkaPub: deps.GetBroker(types.Kafka).GetPublisher(),
		{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub: deps.GetBroker(types.RabbitMQ).GetPublisher(),
	}
	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}

func (uc *{{camel .ModuleName}}UsecaseImpl) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, meta candishared.Meta, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}data, err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().FetchAll(ctx, filter)
	if err != nil {
		return data, meta, err
	}
	count := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Count(ctx, filter)
	meta = candishared.NewMeta(filter.Page, filter.Limit, count){{end}}

	return
}

func (uc *{{camel .ModuleName}}UsecaseImpl) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id string) (data shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: id}
	data, err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter){{end}}
	return
}

func (uc *{{camel .ModuleName}}UsecaseImpl) Create{{upper (camel .ModuleName)}}(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	return{{if or .SQLDeps .MongoDeps}} uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, data){{end}}
}

func (uc *{{camel .ModuleName}}UsecaseImpl) Update{{upper (camel .ModuleName)}}(ctx context.Context, id string, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: id}
	existing, err := uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter)
	if err != nil {
		return err
	}
	data.ID = existing.ID
	data.CreatedAt = existing.CreatedAt
	err = uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, data){{end}}
	return
}

func (uc *{{camel .ModuleName}}UsecaseImpl) Delete{{upper (camel .ModuleName)}}(ctx context.Context, id string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()
	{{if and .MongoDeps (not .SQLDeps)}}
	objID, _ := primitive.ObjectIDFromHex(id){{end}}
	return {{if or .SQLDeps .MongoDeps}}uc.repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.{{upper (camel .ModuleName)}}Repo().Delete(ctx, &shareddomain.{{upper (camel .ModuleName)}}{
		ID: {{if and .MongoDeps (not .SQLDeps)}}objID{{else}}id{{end}},
	}){{end}}
}
`

	templateUsecaseTest = `// {{.Header}}

package usecase

import (
	"context"
	"errors"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"testing"

	mockdeps "{{.LibraryName}}/mocks/codebase/factory/dependency"
	mockinterfaces "{{.LibraryName}}/mocks/codebase/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew{{upper (camel .ModuleName)}}Usecase(t *testing.T) {
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

	uc, setFunc := New{{upper (camel .ModuleName)}}Usecase(mockDeps)
	setFunc(nil)
	assert.NotNil(t, uc)
}

func Test_{{camel .ModuleName}}UsecaseImpl_GetAll{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]shareddomain.{{upper (camel .ModuleName)}}{}, nil)
		{{camel .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		_, _, err := uc.GetAll{{upper (camel .ModuleName)}}(context.Background(), &domain.Filter{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]shareddomain.{{upper (camel .ModuleName)}}{}, errors.New("Error"))
		{{camel .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		_, _, err := uc.GetAll{{upper (camel .ModuleName)}}(context.Background(), &domain.Filter{{upper (camel .ModuleName)}}{})
		assert.Error(t, err)
	})
}

func Test_{{camel .ModuleName}}UsecaseImpl_GetDetail{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		responseData := shareddomain.{{upper (camel .ModuleName)}}{}

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(responseData, nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		result, err := uc.GetDetail{{upper (camel .ModuleName)}}(context.Background(), "id")
		assert.NoError(t, err)
		assert.Equal(t, responseData, result)
	})
}

func Test_{{camel .ModuleName}}UsecaseImpl_Create{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Create{{upper (camel .ModuleName)}}(context.Background(), &shareddomain.{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})
}

func Test_{{camel .ModuleName}}UsecaseImpl_Update{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, nil)
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(context.Background(), "id", &shareddomain.{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, errors.New("Error"))
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(context.Background(), "id", &shareddomain.{{upper (camel .ModuleName)}}{})
		assert.Error(t, err)
	})
}

func Test_{{camel .ModuleName}}UsecaseImpl_Delete{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}{}
		repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}}: repo{{if .SQLDeps}}SQL{{else}}Mongo{{end}},
		}

		err := uc.Delete{{upper (camel .ModuleName)}}(context.Background(), "id")
		assert.NoError(t, err)
	})
}
`
)
