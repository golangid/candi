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
	"{{.LibraryName}}/codebase/factory/dependency"{{if or .KafkaHandler .RabbitMQHandler}}
	"{{.LibraryName}}/codebase/factory/types"{{end}}
	"{{.LibraryName}}/codebase/interfaces"
)

// {{upper (camel .ModuleName)}}Usecase abstraction
type {{upper (camel .ModuleName)}}Usecase interface {
	GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []domain.Response{{upper (camel .ModuleName)}}, meta candishared.Meta, err error)
	GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id string) (data domain.Response{{upper (camel .ModuleName)}}, err error)
	Create{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (err error)
	Update{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (err error)
	Delete{{upper (camel .ModuleName)}}(ctx context.Context, id string) (err error)
}

type {{camel .ModuleName}}UsecaseImpl struct {
	sharedUsecase common.Usecase
	cache         interfaces.Cache
	{{if not .SQLDeps}}// {{end}}repoSQL       repository.RepoSQL
	{{if not .MongoDeps}}// {{end}}repoMongo     repository.RepoMongo{{if .ArangoDeps}}
	repoArango     repository.RepoArango{{end}}
	{{if not .KafkaHandler}}// {{ end }}kafkaPub      interfaces.Publisher
	{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub   interfaces.Publisher
}

// New{{upper (camel .ModuleName)}}Usecase usecase impl constructor
func New{{upper (camel .ModuleName)}}Usecase(deps dependency.Dependency) ({{upper (camel .ModuleName)}}Usecase, func(sharedUsecase common.Usecase)) {
	uc := &{{camel .ModuleName}}UsecaseImpl{
		{{if not .RedisDeps}}// {{end}}cache: deps.GetRedisPool().Cache(),
		{{if not .SQLDeps}}// {{end}}repoSQL:   repository.GetSharedRepoSQL(),
		{{if not .MongoDeps}}// {{end}}repoMongo: repository.GetSharedRepoMongo(),{{if .ArangoDeps}}
		repoArango: repository.GetSharedRepoArango(),{{end}}
		{{if not .KafkaHandler}}// {{ end }}kafkaPub: deps.GetBroker(types.Kafka).GetPublisher(),
		{{if not .RabbitMQHandler}}// {{ end }}rabbitmqPub: deps.GetBroker(types.RabbitMQ).GetPublisher(),
	}
	return uc, func(sharedUsecase common.Usecase) {
		uc.sharedUsecase = sharedUsecase
	}
}
`

	templateUsecaseGetAll = `package usecase

import (
	"context"
	"time"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (results []domain.Response{{upper (camel .ModuleName)}}, meta candishared.Meta, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	var data []shareddomain.{{upper (camel .ModuleName)}}
	{{if or .SQLDeps .MongoDeps .ArangoDeps}}data, err = uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().FetchAll(ctx, filter)
	if err != nil {
		return results, meta, err
	}
	count := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Count(ctx, filter)
	meta = candishared.NewMeta(filter.Page, filter.Limit, count){{end}}

	for _, detail := range data {
		results = append(results, domain.Response{{upper (camel .ModuleName)}}{
			ID: detail.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}},
			Field: detail.Field,
			CreatedAt: detail.CreatedAt.Format(time.RFC3339),
			UpdatedAt: detail.UpdatedAt.Format(time.RFC3339),
		})
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
	"time"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, id string) (result domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	var data shareddomain.{{upper (camel .ModuleName)}}
	{{if or .SQLDeps .MongoDeps .ArangoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: id}
	data, err = uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter){{end}}
	if err != nil {
		return result, err
	}

	result.ID = data.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
	result.Field = data.Field
	result.CreatedAt = data.CreatedAt.Format(time.RFC3339)
	result.UpdatedAt = data.UpdatedAt.Format(time.RFC3339)
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

		_, err := uc.GetDetail{{upper (camel .ModuleName)}}(context.Background(), "id")
		assert.NoError(t, err)
	})
}
`

	templateUsecaseCreate = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"{{if or .SQLDeps .MongoDeps .ArangoDeps}}
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"{{end}}

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Create{{upper (camel .ModuleName)}}(ctx context.Context, req *domain.Request{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	return{{if or .SQLDeps .MongoDeps .ArangoDeps}} uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, &shareddomain.{{upper (camel .ModuleName)}}{
		Field: req.Field,
	}){{end}}
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

		err := uc.Create{{upper (camel .ModuleName)}}(context.Background(), &domain.Request{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})
}
`

	templateUsecaseUpdate = `package usecase

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Update{{upper (camel .ModuleName)}}(ctx context.Context, data *domain.Request{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if or .SQLDeps .MongoDeps .ArangoDeps}}repoFilter := domain.Filter{{upper (camel .ModuleName)}}{ID: data.ID}
	existing, err := uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Find(ctx, &repoFilter)
	if err != nil {
		return err
	}
	existing.Field = data.Field
	err = uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Save(ctx, &existing){{end}}
	return
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
	t.Run("Testcase #1: Positive", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, nil)
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(context.Background(), &domain.Request{{upper (camel .ModuleName)}}{})
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {

		{{camel .ModuleName}}Repo := &mockrepo.{{upper (camel .ModuleName)}}Repository{}
		{{camel .ModuleName}}Repo.On("Find", mock.Anything, mock.Anything).Return(shareddomain.{{upper (camel .ModuleName)}}{}, errors.New("Error"))
		{{camel .ModuleName}}Repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}} := &mocksharedrepo.Repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}{}
		repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.On("{{upper (camel .ModuleName)}}Repo").Return({{camel .ModuleName}}Repo)

		uc := {{camel .ModuleName}}UsecaseImpl{
			repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}: repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}},
		}

		err := uc.Update{{upper (camel .ModuleName)}}(context.Background(), &domain.Request{{upper (camel .ModuleName)}}{})
		assert.Error(t, err)
	})
}
`
	templateUsecaseDelete = `package usecase

import (
	"context"{{if or .SQLDeps .MongoDeps .ArangoDeps}}
	
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"{{end}}{{if and .MongoDeps (not .SQLDeps)}}

	"go.mongodb.org/mongo-driver/bson/primitive"{{end}}

	"{{.LibraryName}}/tracer"
)

func (uc *{{camel .ModuleName}}UsecaseImpl) Delete{{upper (camel .ModuleName)}}(ctx context.Context, id string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}Usecase:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	{{if and .MongoDeps (not .SQLDeps)}}objID, _ := primitive.ObjectIDFromHex(id){{end}}
	return {{if or .SQLDeps .MongoDeps .ArangoDeps}}uc.repo{{if .SQLDeps}}SQL{{else if .MongoDeps}}Mongo{{else if .ArangoDeps}}Arango{{end}}.{{upper (camel .ModuleName)}}Repo().Delete(ctx, &shareddomain.{{upper (camel .ModuleName)}}{
		ID: {{if and .MongoDeps (not .SQLDeps)}}objID{{else}}id{{end}},
	}){{end}}
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

		err := uc.Delete{{upper (camel .ModuleName)}}(context.Background(), "id")
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

	uc, setFunc := New{{upper (camel .ModuleName)}}Usecase(mockDeps)
	setFunc(nil)
	assert.NotNil(t, uc)
}
`
)
