package broker

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	// RedisBrokerKey key constant
	RedisBrokerKey = "dynamic_scheduling"
)

// RedisOptionFunc func type
type RedisOptionFunc func(*RedisBroker)

// RedisSetConfigCommands set config commands
func RedisSetConfigCommands(commands ...string) RedisOptionFunc {
	return func(r *RedisBroker) {
		r.configCommands = commands
	}
}

// RedisSetSubscribeChannels set channels
func RedisSetSubscribeChannels(channels ...string) RedisOptionFunc {
	return func(r *RedisBroker) {
		r.subscribeChannels = channels
	}
}

type RedisBroker struct {
	pool              *redis.Pool
	configCommands    []string
	subscribeChannels []string
}

// NewRedisBroker setup redis for publish message
func NewRedisBroker(pool *redis.Pool, opts ...RedisOptionFunc) *RedisBroker {
	r := &RedisBroker{
		pool: pool,
		// default config
		configCommands:    []string{"SET", "notify-keyspace-events", "Ex"},
		subscribeChannels: []string{"__keyevent@*__:expired"},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// GetConfiguration method, return redis pubsub connection
func (r *RedisBroker) GetConfiguration() interface{} {
	conn := r.pool.Get()

	var commands []interface{}
	for _, cmd := range r.configCommands {
		commands = append(commands, cmd)
	}
	conn.Do("CONFIG", commands...)

	psc := &redis.PubSubConn{Conn: conn}
	commands = []interface{}{}
	for _, cmd := range r.subscribeChannels {
		commands = append(commands, cmd)
	}
	psc.PSubscribe(commands...)
	return psc
}

// GetPublisher method
func (r *RedisBroker) GetPublisher() interfaces.Publisher {
	return r
}

// GetName method
func (r *RedisBroker) GetName() types.Worker {
	return types.RedisSubscriber
}

// Health method
func (r *RedisBroker) Health() map[string]error {
	mErr := make(map[string]error)

	ping := r.pool.Get()
	_, err := ping.Do("PING")
	ping.Close()
	mErr[string(types.RedisSubscriber)] = err

	return mErr
}

// Disconnect method
func (r *RedisBroker) Disconnect(ctx context.Context) error {
	deferFunc := logger.LogWithDefer("redis: closing pool...")
	defer deferFunc()

	return r.pool.Close()
}

// PublishMessage method
func (r *RedisBroker) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis_broker:publish_message")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	if args.Key == "" {
		return errors.New("key cannot empty")
	}

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)
	if args.IsDeleteMessage {
		trace.SetTag("is_delete", args.IsDeleteMessage)
		return r.deleteMessage(ctx, args)
	}
	if err := args.Validate(); err != nil {
		return err
	}
	if args.Delay <= 0 {
		return errors.New("delay cannot empty")
	}

	trace.Log("header", args.Header)
	trace.Log("delay", args.Delay.String())
	trace.Log("message", args.Message)

	conn := r.pool.Get()
	defer conn.Close()

	eventID := uuid.NewString()
	trace.SetTag("event_id", eventID)
	redisMessage, _ := json.Marshal(RedisMessage{
		EventID: eventID, HandlerName: args.Topic, Key: args.Key,
	})
	if _, err := conn.Do("SET", string(redisMessage), 1); err != nil {
		return err
	}
	_, err = conn.Do("EXPIRE", string(redisMessage), int(args.Delay.Seconds()))
	_, err = conn.Do("HSET", RedisBrokerKey, args.Key, args.Message)
	return err
}

// deleteMessage method
func (r *RedisBroker) deleteMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	conn := r.pool.Get()
	trace := tracer.StartTrace(ctx, "redis_broker:delete_message")
	defer func() { conn.Close(); trace.Finish(tracer.FinishWithError(err)) }()

	conn.Do("HDEL", RedisBrokerKey, args.Key)

	b, _ := json.Marshal(RedisMessage{
		HandlerName: args.Topic, Key: args.Key,
	})
	b[len(b)-1] = '*'
	key := string(b)
	var keys []string
	if strings.HasSuffix(key, "*") {
		keys, _ = redis.Strings(conn.Do("KEYS", key))
	}
	if len(keys) == 0 {
		keys = []string{key}
	}
	for _, k := range keys {
		_, err = conn.Do("DEL", k)
	}
	return
}

// RedisMessage messaging model for redis subscriber key
type RedisMessage struct {
	HandlerName string `json:"h"`
	Key         string `json:"key"`
	Message     string `json:"message,omitempty"`
	EventID     string `json:"id,omitempty"`
}

// GenerateKeyDeleteRedisPubSubMessage delete redis key pubsub message pattern
func GenerateKeyDeleteRedisPubSubMessage(topic string, message interface{}) string {
	b, _ := json.Marshal(RedisMessage{
		HandlerName: topic, Message: string(candihelper.ToBytes(message)),
	})
	b[len(b)-1] = '*'
	return string(b)
}

// ParseRedisPubSubKeyTopic parse key to redis message
func ParseRedisPubSubKeyTopic(key []byte) (redisMessage RedisMessage) {
	json.Unmarshal(key, &redisMessage)
	return redisMessage
}
