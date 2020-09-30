package dependency

import (
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
)

// Dependency base
type Dependency interface {
	GetMiddleware() interfaces.Middleware
	GetBroker() interfaces.Broker
	GetSQLDatabase() interfaces.SQLDatabase
	GetMongoDatabase() interfaces.MongoDatabase
	GetRedisPool() interfaces.RedisPool
	GetKey() interfaces.RSAKey
	GetValidator() interfaces.Validator
	GetExtended() map[string]interface{}
}

// Option func type
type Option func(*deps)

type deps struct {
	mw        interfaces.Middleware
	broker    interfaces.Broker
	sqlDB     interfaces.SQLDatabase
	mongoDB   interfaces.MongoDatabase
	redisPool interfaces.RedisPool
	key       interfaces.RSAKey
	validator interfaces.Validator
	extended  map[string]interface{}
}

// SetMiddleware option func
func SetMiddleware(mw interfaces.Middleware) Option {
	return func(d *deps) {
		d.mw = mw
	}
}

// SetBroker option func
func SetBroker(broker interfaces.Broker) Option {
	return func(d *deps) {
		d.broker = broker
	}
}

// SetSQLDatabase option func
func SetSQLDatabase(db interfaces.SQLDatabase) Option {
	return func(d *deps) {
		d.sqlDB = db
	}
}

// SetMongoDatabase option func
func SetMongoDatabase(db interfaces.MongoDatabase) Option {
	return func(d *deps) {
		d.mongoDB = db
	}
}

// SetRedisPool option func
func SetRedisPool(db interfaces.RedisPool) Option {
	return func(d *deps) {
		d.redisPool = db
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

// SetExtended option func
func SetExtended(ext map[string]interface{}) Option {
	return func(d *deps) {
		d.extended = ext
	}
}

// InitDependency constructor
func InitDependency(opts ...Option) Dependency {
	opt := new(deps)

	for _, o := range opts {
		o(opt)
	}

	return opt
}

func (d *deps) GetMiddleware() interfaces.Middleware {
	return d.mw
}
func (d *deps) GetBroker() interfaces.Broker {
	return d.broker
}
func (d *deps) GetSQLDatabase() interfaces.SQLDatabase {
	return d.sqlDB
}
func (d *deps) GetMongoDatabase() interfaces.MongoDatabase {
	return d.mongoDB
}
func (d *deps) GetRedisPool() interfaces.RedisPool {
	return d.redisPool
}
func (d *deps) GetKey() interfaces.RSAKey {
	return d.key
}
func (d *deps) GetValidator() interfaces.Validator {
	return d.validator
}
func (d *deps) GetExtended() map[string]interface{} {
	return d.extended
}
