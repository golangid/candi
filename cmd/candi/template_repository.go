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

	"gorm.io/driver/{{.SQLDriver}}"
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
	{{if .SQLUseGORM}}gormRead, err := gorm.Open({{.SQLDriver}}.New({{.SQLDriver}}.Config{
		Conn: readDB,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	gormWrite, err := gorm.Open({{.SQLDriver}}.New({{.SQLDriver}}.Config{
		Conn: writeDB,
	}), &gorm.Config{SkipDefaultTransaction: true})
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

	tx{{if not .SQLUseGORM}}, err{{end}} := r.writeDB.Begin()` + "{{if .SQLUseGORM}}\n	err = tx.Error{{end}}" + `
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
	}(candishared.SetToContext(ctx, candishared.ContextKeySQLTransaction, tx))

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
	defer trace.Finish()

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
			trace.SetError(err)
			trace.SetError(errAbort)
		} else {
			errCommit := r.writeDB.CommitTransaction(ctx, txID, nil)
			trace.SetError(errCommit)
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
)

// {{upper (camel .ModuleName)}}Repository abstract interface
type {{upper (camel .ModuleName)}}Repository interface {
	FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) ([]shareddomain.{{upper (camel .ModuleName)}}, error)
	Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int
	Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (shareddomain.{{upper (camel .ModuleName)}}, error)
	Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) error
	Delete(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error)
}
`

	templateRepositoryMongoImpl = `// {{.Header}}

package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/tracer"
)

type {{camel .ModuleName}}RepoMongo struct {
	readDB, writeDB *mongo.Database
	collection      string
}

// New{{upper (camel .ModuleName)}}RepoMongo mongo repo constructor
func New{{upper (camel .ModuleName)}}RepoMongo(readDB, writeDB *mongo.Database) {{upper (camel .ModuleName)}}Repository {
	return &{{camel .ModuleName}}RepoMongo{
		readDB: 	readDB, 
		writeDB: 	writeDB, 
		collection: shareddomain.{{upper (camel .ModuleName)}}{}.CollectionName(),
	}
}

func (r *{{camel .ModuleName}}RepoMongo) FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:FetchAll")
	defer func() { trace.SetError(err); trace.Finish() }()

	where := bson.M{}
	trace.SetTag("query", where)

	findOptions := options.Find()
	if len(filter.OrderBy) > 0 {
		findOptions.SetSort(filter)
	}

	if !filter.ShowAll {
		findOptions.SetLimit(int64(filter.Limit))
		findOptions.SetSkip(int64(filter.CalculateOffset()))
	}
	cur, err := r.readDB.Collection(r.collection).Find(ctx, where, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	cur.All(ctx, &data)
	return
}

func (r *{{camel .ModuleName}}RepoMongo) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Find")
	defer func() { trace.SetError(err); trace.Finish() }()

	bsonWhere := make(bson.M)
	if filter.ID != "" {
		bsonWhere["_id"], _ = primitive.ObjectIDFromHex(filter.ID)
	}
	trace.SetTag("query", bsonWhere)

	err = r.readDB.Collection(r.collection).FindOne(ctx, bsonWhere).Decode(&result)
	return
}

func (r *{{camel .ModuleName}}RepoMongo) Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Count")
	defer trace.Finish()

	where := bson.M{}
	count, err := r.readDB.Collection(r.collection).CountDocuments(trace.Context(), where)
	trace.SetError(err)
	return int(count)
}

func (r *{{camel .ModuleName}}RepoMongo) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Save")
	defer func() { trace.SetError(err); trace.Finish() }()
	tracer.Log(ctx, "data", data)

	data.UpdatedAt = time.Now()
	if data.ID{{if and .MongoDeps (not .SQLDeps)}}.IsZero(){{else}} == ""{{end}} {
		data.ID = primitive.NewObjectID(){{if or .SQLDeps (not .MongoDeps)}}.Hex(){{end}}
		data.CreatedAt = time.Now()
		_, err = r.writeDB.Collection(r.collection).InsertOne(ctx, data)
	} else {
		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = r.writeDB.Collection(r.collection).UpdateOne(ctx,
			bson.M{
				"_id": data.ID,
			},
			bson.M{
				"$set": data,
			}, &opt)
	}

	return
}

func (r *{{camel .ModuleName}}RepoMongo) Delete(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoMongo:Delete")
	defer func() { trace.SetError(err); trace.Finish() }()

	_, err = r.writeDB.Collection(r.collection).DeleteOne(ctx, bson.M{"_id": data.ID})
	return
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
	defer func() { trace.SetError(err); trace.Finish() }()

	return
}

func (r *{{camel .ModuleName}}RepoArango) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Find")
	defer func() { trace.SetError(err); trace.Finish() }()

	return
}

func (r *{{camel .ModuleName}}RepoArango) Count(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) int {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Count")
	defer trace.Finish()
	
	var total int

	return total
}

func (r *{{camel .ModuleName}}RepoArango) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Save")
	defer func() { trace.SetError(err); trace.Finish() }()
	tracer.Log(ctx, "data", data)

	return
}

func (r *{{camel .ModuleName}}RepoArango) Delete(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoArango:Delete")
	defer func() { trace.SetError(err); trace.Finish() }()

	return
}
`

	templateRepositorySQLImpl = `// {{.Header}}

package repository

import (
	"context"` + `
	{{if not .SQLUseGORM}}"database/sql"
	"fmt"{{end}}` + `{{if .SQLDeps}}
	"time"{{end}}` + `{{if .SQLDeps}}
	"github.com/google/uuid"{{end}}` + `

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"` +
		`{{if .SQLUseGORM}}

	{{ if .IsMonorepo }}"monorepo/globalshared"{{else}}"{{$.PackagePrefix}}/pkg/shared"{{end}}
	"strings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"{{end}}` + `
)

type {{camel .ModuleName}}RepoSQL struct {
	readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB
}

// New{{upper (camel .ModuleName)}}RepoSQL mongo repo constructor
func New{{upper (camel .ModuleName)}}RepoSQL(readDB, writeDB *{{if .SQLUseGORM}}gorm{{else}}sql{{end}}.DB) {{upper (camel .ModuleName)}}Repository {
	return &{{camel .ModuleName}}RepoSQL{
		readDB, writeDB,
	}
}

func (r *{{camel .ModuleName}}RepoSQL) FetchAll(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (data []shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:FetchAll")
	defer func() { trace.SetError(err); trace.Finish() }()

	if filter.OrderBy == "" {
		filter.OrderBy = ` + `"updated_at"` + `
	}
	
	{{if .SQLUseGORM}}db := {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB)
	
	err = db.Order(clause.OrderByColumn{
		Column: clause.Column{Name: filter.OrderBy},
		Desc:   strings.ToUpper(filter.Sort) == "DESC",
	}).Limit(filter.Limit).Offset(filter.CalculateOffset()).Find(&data).Error
	{{else}}query := fmt.Sprintf("SELECT id, field, created_at, updated_at FROM {{snake .ModuleName}}s ORDER BY %s %s LIMIT %d OFFSET %d",
		filter.OrderBy, filter.Sort, filter.Limit, filter.CalculateOffset())
	trace.Log("query", query)
	rows, err := r.readDB.Query(query)
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

	{{if .SQLUseGORM}}db := {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB)
	
	var total int64
	db.Model(&shareddomain.{{upper (camel .ModuleName)}}{}).Count(&total)
	count = int(total)
	{{else}}r.readDB.QueryRow("SELECT COUNT(*) FROM {{snake .ModuleName}}s").Scan(&count){{end}}
	trace.Log("count", count)
	return
}

func (r *{{camel .ModuleName}}RepoSQL) Find(ctx context.Context, filter *domain.Filter{{upper (camel .ModuleName)}}) (result shareddomain.{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Find")
	defer func() { trace.SetError(err); trace.Finish() }()

	{{if .SQLUseGORM}}db := {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, r.readDB)
	if filter.ID != "" {
		db = db.Where("id = ?", filter.ID)
	}

	err = db.First(&result).Error
	{{else}}err = r.readDB.QueryRow("SELECT id, field, created_at, updated_at FROM {{snake .ModuleName}}s WHERE id={{if eq .SQLDriver "postgres"}}$1{{else}}?{{end}}", filter.ID).
		Scan(&result.ID, &result.Field, &result.CreatedAt, &result.UpdatedAt)
	{{end}}return
}

func (r *{{camel .ModuleName}}RepoSQL) Save(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Save")
	defer func() { trace.SetError(err); trace.Finish() }()
	trace.Log("data", data)

	{{if .SQLUseGORM}}db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	if data.ID == "" {
		data.ID = uuid.NewString()
		err = {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db).Create(data).Error
	} else {
		err = {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db).Save(data).Error
	}
	{{else}}var query string
	var args []interface{}

	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	if data.ID == "" {
		data.ID = uuid.NewString()
		query = "INSERT INTO {{snake .ModuleName}}s (id, field, created_at, updated_at) VALUES ({{if eq .SQLDriver "postgres"}}$1,$2,$3,$4{{else}}?,?,?,?{{end}})"
		args = []interface{}{data.ID, data.Field, data.CreatedAt, data.UpdatedAt}
	} else {
		query = "UPDATE {{snake .ModuleName}}s SET field={{if eq .SQLDriver "postgres"}}$1{{else}}?{{end}} created_at={{if eq .SQLDriver "postgres"}}$2{{else}}?{{end}} updated_at={{if eq .SQLDriver "postgres"}}$3{{else}}?{{end}} WHERE id={{if eq .SQLDriver "postgres"}}$4{{else}}?{{end}}"
		args = []interface{}{data.Field, data.CreatedAt, data.UpdatedAt, data.ID}
	}

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

func (r *{{camel .ModuleName}}RepoSQL) Delete(ctx context.Context, data *shareddomain.{{upper (camel .ModuleName)}}) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}RepoSQL:Delete")
	defer func() { trace.SetError(err); trace.Finish() }()

	{{if .SQLUseGORM}}db := r.writeDB
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*gorm.DB); ok {
		db = tx
	}
	err = {{ if .IsMonorepo }}global{{end}}shared.SetSpanToGorm(ctx, db).Delete(data).Error
	{{else}}var stmt *sql.Stmt
	if tx, ok := candishared.GetValueFromContext(ctx, candishared.ContextKeySQLTransaction).(*sql.Tx); ok {
		stmt, err = tx.PrepareContext(ctx, "DELETE FROM {{snake .ModuleName}}s WHERE id={{if eq .SQLDriver "postgres"}}$1{{else}}?{{end}}")
	} else {
		stmt, err = r.writeDB.PrepareContext(ctx, "DELETE FROM {{snake .ModuleName}}s WHERE id={{if eq .SQLDriver "postgres"}}$1{{else}}?{{end}}")
	}

	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, data.ID)
	{{end}}return
}
`
)
