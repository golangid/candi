package dependency

import (
	"context"
	"log"

	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

const (
	primary = "primary"
)

// Dependency base
type Dependency interface {
	GetMiddleware() interfaces.Middleware
	SetMiddleware(mw interfaces.Middleware)

	GetBroker(types.Worker) interfaces.Broker
	FetchBroker(func(types.Worker, interfaces.Broker))
	AddBroker(brokerType types.Worker, b interfaces.Broker)

	// get primary sql database
	GetSQLDatabase() interfaces.SQLDatabase
	// get primary mongo database
	GetMongoDatabase() interfaces.MongoDatabase
	// get primary redis pool
	GetRedisPool() interfaces.RedisPool

	GetSQLDatabaseByKey(key string) interfaces.SQLDatabase
	GetMongoDatabaseByKey(key string) interfaces.MongoDatabase
	GetRedisPoolByKey(key string) interfaces.RedisPool

	GetKey() interfaces.RSAKey
	SetKey(i interfaces.RSAKey)

	GetValidator() interfaces.Validator
	SetValidator(v interfaces.Validator)

	GetLocker() interfaces.Locker
	SetLocker(v interfaces.Locker)

	GetExtended(key string) any
	AddExtended(key string, value any)

	interfaces.Closer
}

// Option func type
type Option func(*deps)

// SetMiddleware option func
func SetMiddleware(mw interfaces.Middleware) Option {
	return func(d *deps) {
		d.mw = mw
	}
}

// SetBrokers option func
func SetBrokers(brokers map[types.Worker]interfaces.Broker) Option {
	return func(d *deps) {
		d.brokers = brokers
	}
}

// SetSQLDatabase option func, set primary sql database instance
func SetSQLDatabase(db interfaces.SQLDatabase) Option {
	return func(d *deps) {
		d.sqlDB = initEmptyMap(d.sqlDB)
		d.sqlDB[primary] = db
	}
}

// AddSQLDatabase option func, add new another sql database instance
func AddSQLDatabase(key string, db interfaces.SQLDatabase) Option {
	return func(d *deps) {
		d.sqlDB = initEmptyMap(d.sqlDB)
		if _, ok := d.sqlDB[key]; ok {
			log.Panicf("sql db for '%s' has been registered", key)
		}
		d.sqlDB[key] = db
	}
}

// SetMongoDatabase option func, set primary mongo database instance
func SetMongoDatabase(db interfaces.MongoDatabase) Option {
	return func(d *deps) {
		d.mongoDB = initEmptyMap(d.mongoDB)
		d.mongoDB[primary] = db
	}
}

// AddMongoDatabase option func, add new another mongo database instance
func AddMongoDatabase(key string, db interfaces.MongoDatabase) Option {
	return func(d *deps) {
		d.mongoDB = initEmptyMap(d.mongoDB)
		if _, ok := d.mongoDB[key]; ok {
			log.Panicf("mongodb for '%s' has been registered", key)
		}
		d.mongoDB[key] = db
	}
}

// SetRedisPool option func, set primary redis pool instance
func SetRedisPool(db interfaces.RedisPool) Option {
	return func(d *deps) {
		d.redisPool = initEmptyMap(d.redisPool)
		d.redisPool[primary] = db
	}
}

// AddRedisPool option func, add new another redis pool instance
func AddRedisPool(key string, db interfaces.RedisPool) Option {
	return func(d *deps) {
		d.redisPool = initEmptyMap(d.redisPool)
		if _, ok := d.redisPool[key]; ok {
			log.Panicf("redis pool for '%s' has been registered", key)
		}
		d.redisPool[key] = db
	}
}

// SetKey option func
func SetKey(key interfaces.RSAKey) Option {
	return func(d *deps) {
		d.key = key
	}
}

// SetValidator option func
func SetValidator(validator interfaces.Validator) Option {
	return func(d *deps) {
		d.validator = validator
	}
}

// SetLocker option func
func SetLocker(lock interfaces.Locker) Option {
	return func(d *deps) {
		d.locker = lock
	}
}

// SetExtended option func
func SetExtended(ext map[string]any) Option {
	return func(d *deps) {
		d.extended = ext
	}
}

// AddExtended option function for add extended
func AddExtended(key string, value any) Option {
	return func(d *deps) {
		d.extended = initEmptyMap(d.extended)
		d.extended[key] = value
	}
}

func initEmptyMap[K comparable, T any](data map[K]T) map[K]T {
	if data == nil {
		data = make(map[K]T)
	}
	return data
}

func safeClose(ctx context.Context, d interfaces.Closer) error {
	if d != nil {
		return d.Disconnect(ctx)
	}
	return nil
}
