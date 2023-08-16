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
	{{ if not (or .SQLDeps .MongoDeps .ArangoDeps) }}// {{end}}"{{.PackagePrefix}}/pkg/shared/repository"
	"{{$.PackagePrefix}}/pkg/shared/usecase/common"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
)

// {{upper (camel .ModuleName)}}Usecase abstraction
type {{upper (camel .ModuleName)}}Usecase interface {
	GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []domain.Response{{upper (camel .ModuleName)}}, meta candishared.Meta, err error)
	GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}}) (data domain.Response{{upper (camel .ModuleName)}}, err error)
	Create{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (res domain.Response{{upper (camel .ModuleName)}}, err error) 
	Update{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (err error)
	Delete{{upper (camel .ModuleName)}}(ctx context.Context, id {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}}) (err error)
}

type {{camel .ModuleName}}UsecaseImpl struct {
	sharedUsecase common.Usecase
	cache         interfaces.Cache
	locker        interfaces.Locker
	{{if not .SQLDeps}}// {{end}}repoSQL       repository.RepoSQL
	{{if not .MongoDeps}}// {{end}}repoMongo     repository.RepoMongo{{if .ArangoDeps}}
	repoArango     repository.RepoArango{{end}}
	publisher      map[types.Worker]interfaces.Publisher
}

// New{{upper (camel .ModuleName)}}Usecase usecase impl constructor
func New{{upper (camel .ModuleName)}}Usecase(deps dependency.Dependency) ({{upper (camel .ModuleName)}}Usecase, func(sharedUsecase common.Usecase)) {
	uc := &{{camel .ModuleName}}UsecaseImpl{
		{{if not .SQLDeps}}// {{end}}repoSQL:   repository.GetSharedRepoSQL(),
		{{if not .MongoDeps}}// {{end}}repoMongo: repository.GetSharedRepoMongo(),
		locker:    deps.GetLocker(),
		publisher: make(map[types.Worker]interfaces.Publisher),{{if .ArangoDeps}}
		repoArango: repository.GetSharedRepoArango(),{{end}}
	}
	if redisPool := deps.GetRedisPool(); redisPool != nil {
		uc.cache = redisPool.Cache()
	}
	if kafkaBroker := deps.GetBroker(types.Kafka); kafkaBroker != nil {
		uc.publisher[types.Kafka] = kafkaBroker.GetPublisher()
	}{{if .RedisSubsHandler}}
	if redisBroker := deps.GetBroker(types.RedisSubscriber); redisBroker != nil {
		uc.publisher[types.RedisSubscriber] = redisBroker.GetPublisher()
	}{{ end }}{{if .RabbitMQHandler}}
	if rabbitmqBroker := deps.GetBroker(types.RabbitMQ); rabbitmqBroker != nil {
		uc.publisher[types.RabbitMQ] = rabbitmqBroker.GetPublisher()
	}{{ end }}
	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}
`

	templateUsecaseGetAll = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (results []domain.Response{{upper (camel .ModuleName)}}, meta candishared.Meta, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps .ArangoDeps}}data, err := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().FetchAll(ctx, filter)
	if err != nil {
		return results, meta, err
	}
	count := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Count(ctx, filter)
	meta = candishared.NewMeta(filter.Page, filter.Limit, count){{end}}

	for _, detail := range data {
		var res domain.Response{{upper (camel .ModuleName)}}
		res.Serialize(&detail)
		results = append(results, res)
	}

	return
}
`

	templateUsecaseGetAllTest = `package usecase

import (
	"context"
	"errors"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_{{camel .ModuleName}}UsecaseImpl_GetAll{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]shareddomain.{{upper (camel .ModuleName)}}{}, nil)
		{{camel .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		_, _, err := uc.GetAll{{upper (camel .ModuleName)}}(context.Background(), &domain.Filter{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("FetchAll", mock.Anything, mock.Anything, mock.Anything).Return([]shareddomain.{{upper (camel .ModuleName)}}{}, errors.New("Error"))
		{{camel .ModuleName}}Repo.On("Count", mock.Anything, mock.Anything).Return(10)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		_, _, err := uc.GetAll{{upper (camel .ModuleName)}}(context.Background(), &domain.Filter{{upper (camel .ModuleName)}}{})
		assert.Error(t, err)
	})
}
`

	templateUsecaseGetDetail = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}}) (result domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps .ArangoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: &id}
	data, err := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter){{end}}
	if err != nil {
		return result, err
	}

	result.Serialize(&data)
	return
}
`

	templateUsecaseGetDetailTest = `package usecase

import (
	"context"

	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_{{camel .ModuleName}}UsecaseImpl_GetDetail{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		_, err := uc.GetDetail{{upper (camel .ModuleName)}}(context.Background(), {{if and .MongoDeps (not .SQLDeps)}}"1"{{else}}1{{end}})
		assert.NoError(t, err)
	})
}
`

	templateUsecaseCreate = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Create{{upper (camel .ModuleName)}}(ctx context.Context, req *domain.Request{{upper (camel .ModuleName)}}) (result domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	data := req.Deserialize()
	err = {{if or .SQLDeps .MongoDeps .ArangoDeps}}uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, &data){{end}}
	result.Serialize(&data)

	/*
	// Sample using publisher
	uc.publisher[types.Kafka].PublishMessage(ctx, &candishared.PublisherArgument{
		Topic: "[topic]",
		Key:   "[key]",
		Message: candihelper.ToBytes([message]),
	})
	*/
	return
}
`

	templateUsecaseCreateTest = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_{{camel .ModuleName}}UsecaseImpl_Create{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		_, err := uc.Create{{upper (camel .ModuleName)}}(context.Background(), &domain.Request{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})
}
`

	templateUsecaseUpdate = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Update{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps .ArangoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: &data.ID}
	existing, err := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter)
	if err != nil {
		return err
	}
	existing.Field = data.Field
	{{if .SQLDeps}}err = uc.repoSQL.WithTransaction(ctx, func(ctx context.Context) error {
		return uc.repoSQL.{{upper (camel .ModuleName)}}Repo().Save(ctx, &existing, candishared.DBUpdateSetUpdatedFields("Field"))
	}){{else}}
	err = uc.repo{{if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, &existing, candishared.DBUpdateSetUpdatedFields("Field")){{end}}
	return{{end}}
}
`

	templateUsecaseUpdateTest = `package usecase

import (
	"context"
	"errors"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_{{camel .ModuleName}}UsecaseImpl_Update{{upper (camel .ModuleName)}}(t *testing.T) {
	ctx := context.Background()
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, nil)
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything, mock.AnythingOfType("candishared.DBUpdateOptionFunc")).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)
		{{if .SQLDeps}}repoSQL.On("WithTransaction", mock.Anything,
			mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				arg := args.Get(1).(func(context.Context) error)
				arg(ctx)
			}){{end}}
		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(ctx, &domain.Request{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, errors.New("Error"))
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything, mock.AnythingOfType("candishared.DBUpdateOptionFunc")).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)
		{{if .SQLDeps}}repoSQL.On("WithTransaction", mock.Anything,
			mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				arg := args.Get(1).(func(context.Context) error)
				arg(ctx)
			}){{end}}
		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(ctx, &domain.Request{{upper (camel .ModuleName)}}{})
		assert.Error(t, err)
	})
}
`
	templateUsecaseDelete = `package usecase

import (
	"context"{{if or .SQLDeps .MongoDeps .ArangoDeps}}
	
	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"{{end}}

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Delete{{upper (camel .ModuleName)}}(ctx context.Context, id {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: &id}
	return {{if or .SQLDeps .MongoDeps .ArangoDeps}}uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Delete(ctx, &repoFilter){{end}}
}
`

	templateUsecaseDeleteTest = `package usecase

import (
	"context"

	mockrepo "{{$.PackagePrefix}}/pkg/mocks/modules/{{cleanPathModule .ModuleName}}/repository"
	mocksharedrepo "{{$.PackagePrefix}}/pkg/mocks/shared/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_{{camel .ModuleName}}UsecaseImpl_Delete{{upper (camel .ModuleName)}}(t *testing.T) {
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		err := uc.Delete{{upper (camel .ModuleName)}}(context.Background(), {{if and .MongoDeps (not .SQLDeps)}}"1"{{else}}1{{end}})
		assert.NoError(t, err)
	})
}
`

	templateUsecaseTest = `// {{.Header}}

package usecase

import (
	"testing"

	mockdeps "{{.LibraryName}}/mocks/codebase/factory/dependency"
	mockinterfaces "{{.LibraryName}}/mocks/codebase/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew{{upper (camel .ModuleName)}}Usecase(t *testing.T) {
	mockPublisher := &mockinterfaces.Publisher{}
	mockBroker := &mockinterfaces.Broker{}
	mockBroker.On("GetPublisher").Return(mockPublisher)

	mockCache := &mockinterfaces.Cache{}
	mockRedisPool := &mockinterfaces.RedisPool{}
	mockRedisPool.On("Cache").Return(mockCache)

	mockDeps := &mockdeps.Dependency{}
	mockDeps.On("GetRedisPool").Return(mockRedisPool)
	mockDeps.On("GetBroker", mock.Anything).Return(mockBroker)
	mockDeps.On("GetLocker").Return(&mockinterfaces.Locker{})

	uc, setFunc := New{{upper (camel .ModuleName)}}Usecase(mockDeps)
	setFunc(nil)
	assert.NotNil(t, uc)
}
`
)
