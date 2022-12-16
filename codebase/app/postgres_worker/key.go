package postgresworker

import (
	"encoding/json"
)

// PostgresHandlerRouteKey key model
type PostgresHandlerRouteKey struct {
	SourceName string `json:"sourceName"`
	TableName  string `json:"tableName"`
}

// String implement stringer
func (p PostgresHandlerRouteKey) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// CreateHandlerRoute creating key pattern for handler
func CreateHandlerRoute(sourceName, tableName string) string {
	return PostgresHandlerRouteKey{
		SourceName: sourceName, TableName: tableName,
	}.String()
}

// ParseHandlerRoute helper
func ParseHandlerRoute(str string) (sourceName, tableName string) {
	var key PostgresHandlerRouteKey
	err := json.Unmarshal([]byte(str), &key)
	if key.SourceName == "" && err != nil {
		key.TableName = str
	}
	return key.SourceName, key.TableName
}
