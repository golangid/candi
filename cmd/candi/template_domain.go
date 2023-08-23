package main

const (
	templateSharedDomain = `package domain

import (
	"time"` +
		`{{if and .MongoDeps (not .SQLDeps)}}
	"go.mongodb.org/mongo-driver/bson/primitive"{{end}}` + `
)

// {{upper (camel .ModuleName)}} model
type {{upper (camel .ModuleName)}} struct {
	ID         {{if and .MongoDeps (not .SQLDeps)}}primitive.ObjectID{{else}}int{{end}}    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:id;primary_key" {{else}}sql:"id" {{end}}` + `{{if .MongoDeps}}bson:"_id" {{end}}` + `json:"id"` + "`" + `
	Field      string    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:field;type:varchar(255)" {{else}}sql:"field" {{end}}` + `{{if .MongoDeps}}bson:"field" {{end}}` + `json:"field"` + "`" + `
	CreatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:created_at" {{else}}sql:"created_at" {{end}}` + `{{if .MongoDeps}}bson:"created_at" {{end}}` + `json:"created_at"` + "`" + `
	UpdatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:updated_at" {{else}}sql:"updated_at" {{end}}` + `{{if .MongoDeps}}bson:"updated_at" {{end}}` + `json:"updated_at"` + "`" + `
}
{{if .SQLUseGORM}}
// TableName return table name of {{upper (camel .ModuleName)}} model
func ({{upper (camel .ModuleName)}}) TableName() string {
	return "{{snake .ModuleName}}s"
}{{end}}
{{if or .MongoDeps .ArangoDeps}}
// CollectionName return collection name of {{upper (camel .ModuleName)}} model
func ({{upper (camel .ModuleName)}}) CollectionName() string {
	return "{{snake .ModuleName}}s"
}{{end}}
`
	templateModuleDomain = `package domain

import "{{.LibraryName}}/candishared"

// Filter{{upper (camel .ModuleName)}} model
type Filter{{upper (camel .ModuleName)}} struct {
	candishared.Filter
	ID        *{{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}} ` + "`json:\"id\"`" + `
	StartDate string ` + "`json:\"startDate\"`" + `
	EndDate   string ` + "`json:\"endDate\"`{{if .SQLUseGORM}}" + `
	Preloads  []string ` + "`json:\"-\"`" + `{{end}}
}
`
	templateModuleRequestDomain = `package domain

import (
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
)

// Request{{upper (camel .ModuleName)}} model
type Request{{upper (camel .ModuleName)}} struct {
	ID    {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}} ` + "`json:\"id\"`" + `
	Field string ` + "`json:\"field\"`" + `
}

// Deserialize to db model
func (r *Request{{upper (camel .ModuleName)}}) Deserialize() (res shareddomain.{{upper (camel .ModuleName)}}) {
	res.Field = r.Field
	return
}
`
	templateModuleResponseDomain = `package domain

import (
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"time"
)

// Response{{upper (camel .ModuleName)}} model
type Response{{upper (camel .ModuleName)}} struct {
	ID        {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}} ` + "`json:\"id\"`" + `
	Field     string ` + "`json:\"field\"`" + `
	CreatedAt string ` + "`json:\"createdAt\"`" + `
	UpdatedAt string ` + "`json:\"updatedAt\"`" + `
}

// Serialize from db model
func (r *Response{{upper (camel .ModuleName)}}) Serialize(source *shareddomain.{{upper (camel .ModuleName)}}) {
	r.ID = source.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
	r.Field = source.Field
	r.CreatedAt = source.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = source.UpdatedAt.Format(time.RFC3339)
}
`
)
