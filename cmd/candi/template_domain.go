package main

const (
	templateSharedDomain = `// {{.Header}}

package domain

import (
	"time"` +
		`{{if and .MongoDeps (not .SQLDeps)}}
	"go.mongodb.org/mongo-driver/bson/primitive"{{end}}` + `
)

// {{upper (camel .ModuleName)}} model
type {{upper (camel .ModuleName)}} struct {
	ID         {{if and .MongoDeps (not .SQLDeps)}}primitive.ObjectID{{else}}string{{end}}    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:id;type:varchar(255);primary_key" {{end}}` + `{{if .MongoDeps}}bson:"_id" {{end}}` + `json:"id"` + "`" + `
	Field      string    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:field;type:varchar(255)" {{end}}` + `{{if .MongoDeps}}bson:"field" {{end}}` + `json:"field"` + "`" + `
	CreatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:created_at" {{end}}` + `{{if .MongoDeps}}bson:"created_at" {{end}}` + `json:"created_at"` + "`" + `
	UpdatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:updated_at" {{end}}` + `{{if .MongoDeps}}bson:"updated_at" {{end}}` + `json:"updated_at"` + "`" + `
}	
{{if .SQLUseGORM}}
// TableName return table name of {{upper (camel .ModuleName)}} model
func ({{upper (camel .ModuleName)}}) TableName() string {
	return "{{snake .ModuleName}}s"
}{{end}}
{{if .MongoDeps}}
// CollectionName return collection name of {{upper (camel .ModuleName)}} model
func ({{upper (camel .ModuleName)}}) CollectionName() string {
	return "{{snake .ModuleName}}s"
}{{end}}
`
	templateModuleDomain = `// {{.Header}}

package domain

import "{{.LibraryName}}/candishared"

// Filter{{upper (camel .ModuleName)}} model
type Filter{{upper (camel .ModuleName)}} struct {
	candishared.Filter
	ID string
}
`
)
