package main

const (
	templateSharedDomain = `// {{.Header}}

package domain

import (
	"time"` +
		`{{if and .MongoDeps (not .SQLDeps)}}
	"go.mongodb.org/mongo-driver/bson/primitive"{{end}}` + `
)

// {{clean (upper .ModuleName)}} model
type {{clean (upper .ModuleName)}} struct {
	ID         {{if and .MongoDeps (not .SQLDeps)}}primitive.ObjectID{{else}}string{{end}}    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:id;type:varchar(255);primary_key" {{end}}` + `{{if .MongoDeps}}bson:"_id" {{end}}` + `json:"id"` + "`" + `
	Field      string    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:field;type:varchar(255)" {{end}}` + `{{if .MongoDeps}}bson:"field" {{end}}` + `json:"field"` + "`" + `
	CreatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:created_at" {{end}}` + `{{if .MongoDeps}}bson:"created_at" {{end}}` + `json:"created_at"` + "`" + `
	ModifiedAt time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:modified_at" {{end}}` + `{{if .MongoDeps}}bson:"modified_at" {{end}}` + `json:"modified_at"` + "`" + `
}	
{{if .SQLUseGORM}}
// TableName return table name of {{clean (upper .ModuleName)}} model
func ({{clean (upper .ModuleName)}}) TableName() string {
	return "{{clean .ModuleName}}s"
}{{end}}
{{if .MongoDeps}}
// CollectionName return collection name of {{clean (upper .ModuleName)}} model
func ({{clean (upper .ModuleName)}}) CollectionName() string {
	return "{{clean .ModuleName}}s"
}{{end}}
`
	templateModuleDomain = `// {{.Header}}

package domain

import "{{.LibraryName}}/candishared"

// Filter{{clean (upper .ModuleName)}} model
type Filter{{clean (upper .ModuleName)}} struct {
	candishared.Filter
	ID string
}
`
)
