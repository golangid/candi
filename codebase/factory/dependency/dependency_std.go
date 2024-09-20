package dependency

import (
	"context"
	"log"

	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

type deps struct {
	mw      interfaces.Middleware
	brokers map[types.Worker]interfaces.Broker

	sqlDB     map[string]interfaces.SQLDatabase
	mongoDB   map[string]interfaces.MongoDatabase
	redisPool map[string]interfaces.RedisPool

	key       interfaces.RSAKey
	validator interfaces.Validator
	locker    interfaces.Locker
	extended  map[string]any
}

var stdDeps = new(deps)

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
	bk := d.brokers[brokerType]
	if bk == nil {
		log.Printf("\x1b[31;1m[dependency.GetBroker] Broker \"%s\" is not registered in dependency config\x1b[0m\n", string(brokerType))
	}
	return bk
}

func (d *deps) FetchBroker(fn func(types.Worker, interfaces.Broker)) {
	for t, bk := range d.brokers {
		fn(t, bk)
	}
}

func (d *deps) AddBroker(brokerType types.Worker, b interfaces.Broker) {
	d.brokers = initEmptyMap(d.brokers)
	d.brokers[brokerType] = b
}

func (d *deps) GetSQLDatabase() interfaces.SQLDatabase {
	return d.sqlDB[primary]
}

func (d *deps) GetMongoDatabase() interfaces.MongoDatabase {
	return d.mongoDB[primary]
}

func (d *deps) GetRedisPool() interfaces.RedisPool {
	return d.redisPool[primary]
}

func (d *deps) GetSQLDatabaseByKey(key string) interfaces.SQLDatabase {
	db, ok := d.sqlDB[key]
	if !ok {
		log.Panicf("[dependency.GetSQLDatabaseByKey]: key '%s' is not registered in sql db", key)
	}
	return db
}

func (d *deps) GetMongoDatabaseByKey(key string) interfaces.MongoDatabase {
	db, ok := d.mongoDB[key]
	if !ok {
		log.Panicf("[dependency.GetMongoDatabaseByKey]: key '%s' is not registered in mongodb", key)
	}
	return db
}

func (d *deps) GetRedisPoolByKey(key string) interfaces.RedisPool {
	db, ok := d.redisPool[key]
	if !ok {
		log.Panicf("[dependency.GetRedisPoolByKey]: key '%s' is not registered in redis pool", key)
	}
	return db
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

func (d *deps) GetExtended(key string) any {
	return d.extended[key]
}

func (d *deps) AddExtended(key string, value any) {
	d.extended = initEmptyMap(d.extended)
	d.extended[key] = value
}

func (d *deps) Disconnect(ctx context.Context) error {
	for _, bk := range d.brokers {
		safeClose(ctx, bk)
	}
	safeClose(ctx, d.locker)
	for _, sqlDeps := range d.sqlDB {
		safeClose(ctx, sqlDeps)
	}
	for _, mongoDeps := range d.mongoDB {
		safeClose(ctx, mongoDeps)
	}
	for _, redisDeps := range d.redisPool {
		safeClose(ctx, redisDeps)
	}
	for _, ext := range d.extended {
		if cl, ok := ext.(interfaces.Closer); ok {
			safeClose(ctx, cl)
		}
	}
	return nil
}

// GetMiddleware public function for get middleware
func GetMiddleware() interfaces.Middleware {
	return stdDeps.mw
}

// GetBroker public function for get broker
func GetBroker(brokerType types.Worker) interfaces.Broker {
	return stdDeps.GetBroker(brokerType)
}

// GetSQLDatabase public function for get sql database
func GetSQLDatabase() interfaces.SQLDatabase {
	return stdDeps.sqlDB[primary]
}

// GetMongoDatabase public function for get mongo database
func GetMongoDatabase() interfaces.MongoDatabase {
	return stdDeps.mongoDB[primary]
}

// GetRedisPool public function for get redis pool
func GetRedisPool() interfaces.RedisPool {
	return stdDeps.redisPool[primary]
}

// GetKey public function for get key (RSA)
func GetKey() interfaces.RSAKey {
	return stdDeps.key
}

// GetValidator public function for get validator
func GetValidator() interfaces.Validator {
	return stdDeps.validator
}

// GetLocker public function for get validator
func GetLocker() interfaces.Locker {
	return stdDeps.locker
}

// GetExtended public function for get extended
func GetExtended(key string) any {
	return stdDeps.GetExtended(key)
}

// GetSQLDatabaseByKey public function for get sql db by key
func GetSQLDatabaseByKey(key string) interfaces.SQLDatabase {
	return stdDeps.GetSQLDatabaseByKey(key)
}

// GetMongoDatabaseByKey public function for get mongo db by key
func GetMongoDatabaseByKey(key string) interfaces.MongoDatabase {
	return stdDeps.GetMongoDatabaseByKey(key)
}

// GetRedisPoolByKey public function for get redis pool by key
func GetRedisPoolByKey(key string) interfaces.RedisPool {
	return stdDeps.GetRedisPoolByKey(key)
}
