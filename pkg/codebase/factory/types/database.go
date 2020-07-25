package types

// Database enum
type Database int

const (
	// SQL database type
	SQL Database = iota

	// Mongo database type
	Mongo

	// Redis database type
	Redis

	// ElasticSearch database type
	ElasticSearch
)
