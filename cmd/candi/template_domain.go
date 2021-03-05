package main

const (
	templateDomain = `// {{.Header}}

package domain

import "time"

// {{clean (upper .ModuleName)}} model
type {{clean (upper .ModuleName)}} struct {
	ID         string    ` + "`" + `{{if .SQLUseGORM}}gorm:"column:id" {{end}}` + `{{if .MongoDeps}}bson:"_id"{{end}} ` + `json:"id"` + "`" + `
	CreatedAt  time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:created_at" {{end}}` + `{{if .MongoDeps}}bson:"created_at"{{end}} ` + `json:"created_at"` + "`" + `
	ModifiedAt time.Time ` + "`" + `{{if .SQLUseGORM}}gorm:"column:modified_at" {{end}}` + `{{if .MongoDeps}}bson:"modified_at"{{end}} ` + `json:"modified_at"` + "`" + `
}	
`
)
