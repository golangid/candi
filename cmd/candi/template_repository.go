package main

const (
	templateRepository = `// {{.Header}} DO NOT EDIT.

package repository

import (
	"sync"
	
	{{ if .ArangoDeps}} arango "github.com/golangid/candi-plugin/arangodb-adapter" {{ end }}

	"{{.LibraryName}}/codebase/factory/dependency"
)

var (
	once sync.Once
)

// SetSharedRepository set the global singleton "RepoSQL" and "RepoMongo" implementation
func SetSharedRepository(deps dependency.Dependency) {
	once.Do(func() {
		{{if not .SQLDeps}}// {{end}}setSharedRepoSQL(deps.GetSQLDatabase().ReadDB(), deps.GetSQLDatabase().WriteDB())
		{{if not .MongoDeps}}// {{end}}setSharedRepoMongo(deps.GetMongoDatabase().ReadDB(), deps.GetMongoDatabase().WriteDB())
		{{if not .ArangoDeps}}// {{end}}setSharedRepoArango(deps.GetExtended("arangodb").(arango.ArangoDatabase).ReadDB(), deps.GetExtended("arangodb").(arango.ArangoDatabase).WriteDB())
	})
}
`

	templateRepositoryUOWSQL = `// {{.Header}}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	// @candi:repositoryImport

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"` +
		`{{if .SQLUseGORM}}

	{{ if .IsMonorepo }}"monorepo/globalshared"{{else}}"{{$.PackagePrefix}}/pkg/shared"{{end}}

	{{if eq .SQLDriver "sqlite3"}}"gorm.io/driver/sqlite"{{else}}"gorm.io/driver/{{.SQLDriver}}"{{end}}
	"gorm.io/gorm"{{end}}` + `
)

type (
	// RepoSQL abstraction
	RepoSQL interface {
		WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) (err error)

		// @candi:repositoryMethod
	}

	repoSQLImpl struct {
		readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB` + "{{if not .SQLUseGORM}}\n		tx    *sql.Tx{{end}}" + `
	
		// register all repository from modules
		// @candi:repositoryField
	}
)

var (
	globalRepoSQL RepoSQL
)

// setSharedRepoSQL set the global singleton "RepoSQL" implementation
func setSharedRepoSQL(readDB, writeDB *sql.DB) {
	{{if .SQLUseGORM}}gormRead, err := gorm.Open({{if eq .SQLDriver "sqlite3"}}sqlite.Dialector{Conn: readDB}{{else}}{{ .SQLDriver }}.New({{ .SQLDriver }}.Config{
		Conn: readDB,
	}){{end}}, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	gormWrite, err := gorm.Open({{if eq .SQLDriver "sqlite3"}}sqlite.Dialector{Conn: writeDB}{{else}}{{ .SQLDriver }}.New({{ .SQLDriver }}.Config{
		Conn: writeDB,
	}){{end}}, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		panic(err)
	}

	{{ if .IsMonorepo }}global{{end}}shared.AddGormCallbacks(gormRead)
	{{ if .IsMonorepo }}global{{end}}shared.AddGormCallbacks(gormWrite){{end}}

	globalRepoSQL = NewRepositorySQL({{if .SQLUseGORM}}gormRead, gormWrite{{else}}readDB, writeDB{{end}})
}

// GetSharedRepoSQL returns the global singleton "RepoSQL" implementation
func GetSharedRepoSQL() RepoSQL {
	return globalRepoSQL
}

// NewRepositorySQL constructor
func NewRepositorySQL(readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB) RepoSQL {

	return &repoSQLImpl{
		readDB: readDB, writeDB: writeDB,

		// @candi:repositoryConstructor
	}
}

// WithTransaction run transaction for each repository with context, include handle canceled or timeout context
func (r *repoSQLImpl) WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "RepoSQL:Transaction")
	defer trace.Finish()

	{{if .SQLUseGORM}}tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB)
	if !ok {
		tx = r.writeDB.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		defer func() {
			if err != nil {
				tx.Rollback()
				trace.SetError(err)
			} else {
				tx.Commit()
			}
		}()
		ctx = candishared.SetToContext(ctx, candishared.ContextKeySQLTransaction, tx)
	}{{else}}tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*sql.Tx)
	if !ok {
		tx, err = r.writeDB.Begin()
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				tx.Rollback()
				trace.SetError(err)
			} else {
				tx.Commit()
			}
		}()
		ctx = candishared.SetToContext(ctx, candishared.ContextKeySQLTransaction, tx)
	}{{end}}

	errChan := make(chan error)
	go func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
			close(errChan)
		}()

		if err := txFunc(ctx); err != nil {
			errChan <- err
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return fmt.Errorf("Canceled or timeout: %v", ctx.Err())
	case e := <-errChan:
		return e
	}
}

// @candi:repositoryImplementation
`

	templateRepositoryUOWMongo = `// {{.Header}} DO NOT EDIT.

package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

	// @candi:repositoryImport
)

type (
	// RepoMongo abstraction
	RepoMongo interface {
		// @candi:repositoryMethod
	}

	repoMongoImpl struct {
		readDB, writeDB *mongo.Database

		// register all repository from modules
		// @candi:repositoryField
	}
)

var globalRepoMongo RepoMongo

// setSharedRepoMongo set the global singleton "RepoMongo" implementation
func setSharedRepoMongo(readDB, writeDB *mongo.Database) {
	globalRepoMongo = &repoMongoImpl{
		readDB: readDB, writeDB: writeDB,

		// @candi:repositoryConstructor
	}
}

// GetSharedRepoMongo returns the global singleton "RepoMongo" implementation
func GetSharedRepoMongo() RepoMongo {
	return globalRepoMongo
}

// @candi:repositoryImplementation
`

	templateRepositoryUOWArango = `// {{.Header}} DO NOT EDIT.

package repository

import (
	"github.com/arangodb/go-driver"

	// @candi:repositoryImport
)

type (
	// RepoArango abstraction
	RepoArango interface {
		WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error, models ...ArangoModels) (err error)

		// @candi:repositoryMethod
	}

	repoArangoImpl struct {
		readDB, writeDB driver.Database

		// register all repository from modules
		// @candi:repositoryField
	}

	ArangoModels interface {
		CollectionName() string
	}
)

var globalRepoArango RepoArango

// setSharedRepoArango set the global singleton "RepoArango" implementation
func setSharedRepoArango(readDB, writeDB driver.Database) {
	globalRepoArango = &repoArangoImpl{
		readDB: readDB, writeDB: writeDB,

		// @candi:repositoryConstructor
	}
}

// GetSharedRepoArango returns the global singleton "RepoArango" implementation
func GetSharedRepoArango() RepoArango{
	return globalRepoArango
}

// @candi:repositoryImplementation

// WithTransaction run transaction for each repository with context, include handle canceled or timeout context
func (r *repoArangoImpl) WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error, models ...ArangoModels) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "RepoSQL:Transaction")
	defer trace.Finish(tracer.FinishWithError(err))

	var cols []string
	for _, value := range models {
		cols = append(cols, value.CollectionName())
	}

	trxCollection := driver.TransactionCollections{
		Read:      cols,
		Write:     cols,
		Exclusive: cols,
	}

	txID, err := r.writeDB.BeginTransaction(ctx, trxCollection, nil)
	if err != nil {
		return err
	}
	trace.Log("transactionId", txID)
	txCtx := driver.WithTransactionID(ctx, txID)

	defer func() {
		if err != nil {
			errAbort := r.writeDB.AbortTransaction(ctx, txID, nil)
			trace.Log("transaction", fmt.Sprintf("transactionId %s is aborted", txID))
			if errAbort != nil {
				trace.Log("transaction_abort", fmt.Sprintf("transactionId %s err %s", txID, errAbort.Error()))
			}
		} else {
			errCommit := r.writeDB.CommitTransaction(ctx, txID, nil)
			if errCommit == nil {
				trace.Log("transaction", fmt.Sprintf("transactionId %s is committed", txID))
			}
			err = errCommit
		}
	}()

	errChan := make(chan error)
	go func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
			close(errChan)
		}()

		if err := txFunc(ctx); err != nil {
			errChan <- err
		}
	}(txCtx)

	select {
	case <-txCtx.Done():
		return fmt.Errorf("Canceled or timeout: %v", txCtx.Err())
	case e := <-errChan:
		return e
	}
}

`

	templateRepositoryAbstraction = `// {{.Header}}

package repository

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
)

// {{upper (camel .ModuleName)}}Repository abstract interface
type {{upper (camel .ModuleName)}}Repository interface {
	FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) ([]shareddomain.{{upper (camel .ModuleName)}}, error)
	Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int
	Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (shareddomain.{{upper (camel .ModuleName)}}, error)
	Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}, updateOptions ...candishared.DBUpdateOptionFunc) error
	Delete(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (err error)
}
`

	templateRepositoryMongoImpl = `// {{.Header}}

package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"{{if not .SQLDeps}}
	"go.mongodb.org/mongo-driver/bson/primitive"{{end}}
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"
)

type {{camel .ModuleName}}RepoMongo struct {
	readDB, writeDB *mongo.Database
	collection      string
	updateTools     *candishared.DBUpdateTools
}

// New{{upper (camel .ModuleName)}}RepoMongo mongo repo constructor
func New{{upper (camel .ModuleName)}}RepoMongo(readDB, writeDB *mongo.Database) {{upper (camel .ModuleName)}}Repository {
	return &{{camel .ModuleName}}RepoMongo{
		readDB: 	readDB, 
		writeDB: 	writeDB, 
		collection: shareddomain.{{upper (camel .ModuleName)}}{}.CollectionName(),
		updateTools: &candishared.DBUpdateTools{
			KeyExtractorFunc: candishared.DBUpdateMongoExtractorKey,
			IgnoredFields:    []string{"_id"},
		},
	}
}

func (r *{{camel .ModuleName}}RepoMongo) FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:FetchAll")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	query := r.setFilter{{upper (camel .ModuleName)}}(filter)
	trace.Log("query", query)

	findOptions := options.Find()
	sort := bson.M{}
	if len(filter.OrderBy) == 0 {
		filter.OrderBy = "updated_at"
	}
	sort[filter.OrderBy] = -1
	if filter.Sort == "ASC" {
		sort[filter.OrderBy] = 1
	}
	findOptions.SetSort(sort)

	if !filter.ShowAll {
		findOptions.SetLimit(int64(filter.Limit))
		findOptions.SetSkip(int64(filter.CalculateOffset()))
	}
	cur, err := r.readDB.Collection(r.collection).Find(ctx, query, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &data)
	return
}

func (r *{{camel .ModuleName}}RepoMongo) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Find")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	query := r.setFilter{{upper (camel .ModuleName)}}(filter)
	trace.Log("query", query)

	err = r.readDB.Collection(r.collection).FindOne(ctx, query).Decode(&result)
	return
}

func (r *{{camel .ModuleName}}RepoMongo) Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Count")
	defer trace.Finish()

	query := r.setFilter{{upper (camel .ModuleName)}}(filter)
	trace.Log("query", query)

	count, err := r.readDB.Collection(r.collection).CountDocuments(ctx, query)
	trace.SetError(err)
	return int(count)
}

func (r *{{camel .ModuleName}}RepoMongo) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}, updateOptions ...candishared.DBUpdateOptionFunc) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Save")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	data.UpdatedAt = time.Now()
	if data.ID{{if and .MongoDeps (not .SQLDeps)}}.IsZero(){{else}} == 0{{end}} {
		data.ID = {{if and .MongoDeps (not .SQLDeps)}}primitive.NewObjectID(){{else}}r.Count(ctx, &domain.Filter{{upper (camel .ModuleName)}}{}) + 1{{end}}
		data.CreatedAt = time.Now()
		_, err = r.writeDB.Collection(r.collection).InsertOne(ctx, data)
		trace.Log("data", data)

	} else {
		updated := bson.M(r.updateTools.ToMap(data, updateOptions...))
		trace.Log("updated", updated)
		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = r.writeDB.Collection(r.collection).UpdateOne(ctx,
			bson.M{
				"_id": data.ID,
			},
			bson.M{
				"$set": updated,
			}, &opt)
	}

	trace.SetTag("id", data.ID.Hex())
	return
}

func (r *{{camel .ModuleName}}RepoMongo) Delete(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Delete")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	_, err = r.writeDB.Collection(r.collection).DeleteOne(ctx, r.setFilter{{upper (camel .ModuleName)}}(filter))
	return
}

func (r *{{camel .ModuleName}}RepoMongo) setFilter{{upper (camel .ModuleName)}}(filter *domain.Filter{{upper (camel .ModuleName)}}) bson.M {

	query := make(bson.M)

	if filter.ID != nil {
		{{if not .SQLDeps}}query["_id"], _ = primitive.ObjectIDFromHex(*filter.ID){{else}}query["_id"] = *filter.ID{{end}}
	}
	if filter.Search != "" {
		query["field"] = bson.M{"$regex": filter.Search}
	}

	return query
}
`

	templateRepositoryArangoImpl = `// {{.Header}}

package repository

import (
	"context"

	"github.com/arangodb/go-driver"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/tracer"
)

type {{camel .ModuleName}}RepoArango struct {
	readDB, writeDB driver.Database
	collection      string
}

// New{{upper (camel .ModuleName)}}RepoArango arango repo constructor
func New{{upper (camel .ModuleName)}}RepoArango(readDB, writeDB driver.Database) {{upper (camel .ModuleName)}}Repository {
	return &{{camel .ModuleName}}RepoArango{
		readDB: 	readDB, 
		writeDB: 	writeDB, 
		collection: shareddomain.{{upper (camel .ModuleName)}}{}.CollectionName(),
	}
}

func (r *{{camel .ModuleName}}RepoArango) FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:FetchAll")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return
}

func (r *{{camel .ModuleName}}RepoArango) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Find")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return
}

func (r *{{camel .ModuleName}}RepoArango) Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Count")
	defer trace.Finish()
	
	var total int

	return total
}

func (r *{{camel .ModuleName}}RepoArango) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}, updateOptions ...candishared.DBUpdateOptionFunc) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Save")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()
	tracer.Log(ctx, "data", data)

	return
}

func (r *{{camel .ModuleName}}RepoArango) Delete(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Delete")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	return
}
`

	templateRepositorySQLImpl = `// {{.Header}}

package repository

import (
	"context"` + `
	{{if not .SQLUseGORM}}"database/sql"
	"fmt"{{end}}` + `{{if .SQLDeps}}
	"time"{{end}}` + `{{if not .SQLUseGORM}}
	"errors"{{end}}
	"strings"` + `

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"` +
		`{{if .SQLUseGORM}}

	{{ if .IsMonorepo }}"monorepo/globalshared"{{else}}"{{$.PackagePrefix}}/pkg/shared"{{end}}
	"gorm.io/gorm"
	"gorm.io/gorm/clause"{{end}}` + `
)

type {{camel .ModuleName}}RepoSQL struct {
	readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB
	updateTools     *candishared.DBUpdateTools
}

// New{{upper (camel .ModuleName)}}RepoSQL mongo repo constructor
func New{{upper (camel .ModuleName)}}RepoSQL(readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB) {{upper (camel .ModuleName)}}Repository {
	return &{{camel .ModuleName}}RepoSQL{
		readDB: readDB, writeDB: writeDB,
		updateTools: &candishared.DBUpdateTools{
			{{if .SQLUseGORM}}KeyExtractorFunc: candishared.DBUpdateGORMExtractorKey, {{end}}IgnoredFields: []string{"id"},
		},
	}
}

func (r *{{camel .ModuleName}}RepoSQL) FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:FetchAll")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	if filter.OrderBy == "" {
		filter.OrderBy = ` + `"updated_at"` + `
	}

	{{if .SQLUseGORM}}db := r.setFilter{{upper (camel .ModuleName)}}({{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB), filter).Order(clause.OrderByColumn{
		Column: clause.Column{Name: filter.OrderBy},
		Desc:   strings.ToUpper(filter.Sort) == "DESC",
	})
	if filter.Limit >= 0 {
		db = db.Limit(filter.Limit).Offset(filter.CalculateOffset())
	}
	err = db.Find(&data).Error
	{{else}}where, args := r.setFilter{{upper (camel .ModuleName)}}(filter)
	if len(args) > 0 {
		where = " WHERE " + where
	}
	query := fmt.Sprintf("SELECT id, field, created_at, updated_at FROM {{snake .ModuleName}}s%s ORDER BY %s %s LIMIT %d OFFSET %d",
		where, filter.OrderBy, filter.Sort, filter.Limit, filter.CalculateOffset())
	trace.Log("query", query)
	rows, err := r.readDB.Query(query, args...)
	if err != nil {
		return data, err
	}
	defer rows.Close()
	for rows.Next() {
		var res shareddomain.{{upper (camel .ModuleName)}}
		if err := rows.Scan(&res.ID, &res.Field, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, err
		}
		data = append(data, res)
	}
	{{end}}return
}

func (r *{{camel .ModuleName}}RepoSQL) Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (count int) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Count")
	defer trace.Finish()

	{{if .SQLUseGORM}}var total int64
	r.setFilter{{upper (camel .ModuleName)}}({{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB), filter).Model(&shareddomain.{{upper (camel .ModuleName)}}{}).Count(&total)
	count = int(total)
	{{else}}where, args := r.setFilter{{upper (camel .ModuleName)}}(filter)
	if len(args) > 0 {
		where = " WHERE " + where
	}
	query := "SELECT COUNT(*) FROM {{snake .ModuleName}}s" + where
	r.readDB.QueryRow(query, args...).Scan(&count)
	trace.Log("query", query)
	trace.Log("args", args){{end}}
	trace.Log("count", count)
	return
}

func (r *{{camel .ModuleName}}RepoSQL) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Find")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	{{if .SQLUseGORM}}err = r.setFilter{{upper (camel .ModuleName)}}({{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB), filter).First(&result).Error
	{{else}}where, args := r.setFilter{{upper (camel .ModuleName)}}(filter)
	query := "SELECT id, field, created_at, updated_at FROM {{snake .ModuleName}}s WHERE " + where + " LIMIT 1"
	trace.Log("query", query)
	trace.Log("args", args)
	err = r.readDB.QueryRow(query, args...).
		Scan(&result.ID, &result.Field, &result.CreatedAt, &result.UpdatedAt)
	{{end}}return
}

func (r *{{camel .ModuleName}}RepoSQL) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}, updateOptions ...candishared.DBUpdateOptionFunc) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Save")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	{{if .SQLUseGORM}}db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	if data.ID == 0 {
		err = {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db).Omit(clause.Associations).Create(data).Error
	} else {
		err = {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db).Model(data).Omit(clause.Associations).Updates(r.updateTools.ToMap(data, updateOptions...)).Error
	}
	{{else}}var query string
	var args []interface{}

	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	if data.ID == 0 {
		query = "INSERT INTO {{snake .ModuleName}}s (field, created_at, updated_at) VALUES ({{if eq .SQLDriver "postgres"}}$1,$2,$3{{else}}?,?,?{{end}})"
		args = []interface{}{data.Field, data.CreatedAt, data.UpdatedAt}
	} else {
		var updatedFields []string{{if eq .SQLDriver "postgres"}}
		i := 1{{end}}
		for field, val := range r.updateTools.ToMap(data, updateOptions...) {
			args = append(args, val)
			updatedFields = append(updatedFields, {{if eq .SQLDriver "postgres"}}fmt.Sprintf("%s=$%d", field, i))
			i++{{else}}fmt.Sprintf("%s=?", field)){{end}}
		}
		query = fmt.Sprintf("UPDATE {{snake .ModuleName}}s SET %s WHERE id={{if eq .SQLDriver "postgres"}}$%d", strings.Join(updatedFields, ", "), i){{else}}?", strings.Join(updatedFields, ", ")){{end}}
		args = append(args, data.ID)
	}
	trace.Log("query", query)
	trace.Log("args", args)

	var stmt *sql.Stmt
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*sql.Tx); ok {
		stmt, err = tx.PrepareContext(ctx, query)
	} else {
		stmt, err = r.writeDB.PrepareContext(ctx, query)
	}

	if err != nil {
		return err
	}
	sqlRes, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	id, _ := sqlRes.LastInsertId()
	data.ID = int(id)
	{{end}}return
}

func (r *{{camel .ModuleName}}RepoSQL) Delete(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Delete")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	{{if .SQLUseGORM}}db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	err = r.setFilter{{upper (camel .ModuleName)}}({{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db), filter).Delete(&shareddomain.{{upper (camel .ModuleName)}}{}).Error
	{{else}}where, args := r.setFilter{{upper (camel .ModuleName)}}(filter)
	if len(args) == 0 {
		return errors.New("Cannot empty filter")
	}
	query :=  "DELETE FROM {{snake .ModuleName}}s WHERE " + where
	trace.Log("query", query)
	trace.Log("args", args)
	var stmt *sql.Stmt
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*sql.Tx); ok {
		stmt, err = tx.PrepareContext(ctx, query)
	} else {
		stmt, err = r.writeDB.PrepareContext(ctx, query)
	}

	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, args...)
	{{end}}return
}

func (r *{{camel .ModuleName}}RepoSQL) setFilter{{upper (camel .ModuleName)}}({{if .SQLUseGORM}}db *gorm.DB, {{end}}filter *domain.Filter{{upper (camel .ModuleName)}}) {{if .SQLUseGORM}}*gorm.DB{{else}}(query string, args []interface{}){{end}} {
	{{if not .SQLUseGORM}}
	var wheres []string{{end}}
	if filter.ID != nil {
		{{if .SQLUseGORM}}db = db.Where("id = ?", *filter.ID){{else}}wheres = append(wheres, {{if eq .SQLDriver "postgres"}}fmt.Sprintf("id = $%d", len(args)+1){{else}}"id = ?"{{end}})
		args = append(args, *filter.ID){{end}}
	}
	if filter.Search != "" {
		{{if .SQLUseGORM}}db = db.Where("(field ILIKE '%%' || ? || '%%')", filter.Search){{else}}wheres = append(wheres, {{if eq .SQLDriver "postgres"}}fmt.Sprintf("field ILIKE '%%%%' || $%d || '%%%%'", len(args)+1){{else}}"field ILIKE '%%' || ? || '%%'"{{end}})
		args = append(args, filter.Search){{end}}
	}{{if .SQLUseGORM}}

	for _, preload := range filter.Preloads {
		db = db.Preload(preload)
	}{{end}}

	return {{if .SQLUseGORM}}db{{else}}strings.Join(wheres, " AND "), args{{end}}
}
`
)
