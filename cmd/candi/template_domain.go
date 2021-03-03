package main

const (
	templateDomain = `// {{.Header}}

package domain

import "time"

// {{clean (upper .ModuleName)}} model
type {{clean (upper .ModuleName)}} struct {
	ID         string    ` + "`" + `json:"id"` + "`" + `
	CreatedAt  time.Time ` + "`" + `json:"created_at"` + "`" + `
	ModifiedAt time.Time ` + "`" + `json:"modified_at"` + "`" + `
}	
`
)
