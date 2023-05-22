package redisworker

import (
	"encoding/json"

	"github.com/golangid/candi/candihelper"
	"github.com/google/uuid"
)

// Deprecated, move to broker package

// RedisMessage model for redis subscriber key
type RedisMessage struct {
	HandlerName string `json:"h"`
	Message     string `json:"message"`
	EventID     string `json:"id,omitempty"`
}

// CreateRedisPubSubMessage create new redis pubsub message
func CreateRedisPubSubMessage(topic string, message interface{}) string {
	key, _ := json.Marshal(RedisMessage{
		EventID: uuid.NewString(), HandlerName: topic, Message: string(candihelper.ToBytes(message)),
	})
	return string(key)
}

// DeleteRedisPubSubMessage delete redis key pubsub message pattern
func DeleteRedisPubSubMessage(topic string, message interface{}) string {
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
