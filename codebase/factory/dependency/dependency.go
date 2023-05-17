package dependency

import (
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

// Dependency base
type Dependency interface {
	GetMiddleware() interfaces.Middleware
	SetMiddleware(mw interfaces.Middleware)

	GetBroker(types.Worker) interfaces.Broker
	AddBroker(brokerType types.Worker, b interfaces.Broker)

	GetSQLDatabase() interfaces.SQLDatabase
	GetMongoDatabase() interfaces.MongoDatabase
	GetRedisPool() interfaces.RedisPool

	GetKey() interfaces.RSAKey
	SetKey(i interfaces.RSAKey)

	GetValidator() interfaces.Validator
	SetValidator(v interfaces.Validator)

	GetLocker() interfaces.Locker
	SetLocker(v interfaces.Locker)

	GetExtended(key string) interface{}
	AddExtended(key string, value interface{})
}

// Option func type
type Option func(*deps)

type deps struct {
	mw        interfaces.Middleware
	brokers   map[types.Worker]interfaces.Broker
	sqlDB     interfaces.SQLDatabase
	mongoDB   interfaces.MongoDatabase
	redisPool interfaces.RedisPool
	key       interfaces.RSAKey
	validator interfaces.Validator
	locker    interfaces.Locker
	extended  map[string]interface{}
}

var stdDeps = new(deps)

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

// SetLocker option func
func SetLocker(lock interfaces.Locker) Option {
	return func(d *deps) {
		d.locker = lock
	}
}

// SetExtended option func
func SetExtended(ext map[string]interface{}) Option {
	return func(d *deps) {
		d.extended = ext
	}
}

// AddExtended option function for add extended
func AddExtended(key string, value interface{}) Option {
	return func(d *deps) {
		if d.extended == nil {
			d.extended = make(map[string]interface{})
		}
		d.extended[key] = value
	}
}

// InitDependency constructor
func InitDependency(opts ...Option) Dependency {
	for _, o := range opts {
		o(stdDeps)
	}

	return stdDeps
}

func (d *deps) GetMiddleware() interfaces.Middleware {
	return d.mw
}
func (d *deps) SetMiddleware(mw interfaces.Middleware) {
	d.mw = mw
}
func (d *deps) GetBroker(brokerType types.Worker) interfaces.Broker {
	return d.brokers[brokerType]
}
func (d *deps) AddBroker(brokerType types.Worker, b interfaces.Broker) {
	if d.brokers == nil {
		d.brokers = make(map[types.Worker]interfaces.Broker)
	}
	d.brokers[brokerType] = b
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
func (d *deps) SetKey(i interfaces.RSAKey) {
	d.key = i
}
func (d *deps) GetValidator() interfaces.Validator {
	return d.validator
}
func (d *deps) SetValidator(v interfaces.Validator) {
	d.validator = v
}
func (d *deps) GetLocker() interfaces.Locker {
	return d.locker
}
func (d *deps) SetLocker(v interfaces.Locker) {
	d.locker = v
}
func (d *deps) GetExtended(key string) interface{} {
	return d.extended[key]
}
func (d *deps) AddExtended(key string, value interface{}) {
	if d.extended == nil {
		d.extended = make(map[string]interface{})
	}
	d.extended[key] = value
}

// GetMiddleware free function for get middleware
func GetMiddleware() interfaces.Middleware {
	return stdDeps.mw
}

// GetBroker free function for get broker
func GetBroker(brokerType types.Worker) interfaces.Broker {
	return stdDeps.brokers[brokerType]
}

// GetSQLDatabase free function for get sql database
func GetSQLDatabase() interfaces.SQLDatabase {
	return stdDeps.sqlDB
}

// GetMongoDatabase free function for get mongo database
func GetMongoDatabase() interfaces.MongoDatabase {
	return stdDeps.mongoDB
}

// GetRedisPool free function for get redis pool
func GetRedisPool() interfaces.RedisPool {
	return stdDeps.redisPool
}

// GetKey free function for get key (RSA)
func GetKey() interfaces.RSAKey {
	return stdDeps.key
}

// GetValidator free function for get validator
func GetValidator() interfaces.Validator {
	return stdDeps.validator
}

// GetLocker free function for get validator
func GetLocker() interfaces.Locker {
	return stdDeps.locker
}

// GetExtended free function for get extended
func GetExtended(key string) interface{} {
	return stdDeps.extended[key]
}
