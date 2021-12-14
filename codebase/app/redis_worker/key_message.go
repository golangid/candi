package redisworker

import (
	"encoding/json"

	"github.com/golangid/candi/candihelper"
	"github.com/google/uuid"
)

// RedisMessage model for redis subscriber key
type RedisMessage struct {
	EventID     string `json:"id"`
	HandlerName string `json:"h"`
	Message     string `json:"message"`
}

// CreateRedisPubSubMessage create new redis pubsub message
func CreateRedisPubSubMessage(topic string, message interface{}) string {
	key, _ := json.Marshal(RedisMessage{
		EventID: uuid.NewString(), HandlerName: topic, Message: string(candihelper.ToBytes(message)),
	})
	return string(key)
}

// ParseRedisPubSubKeyTopic parse key to redis message
func ParseRedisPubSubKeyTopic(key []byte) (redisMessage RedisMessage) {
	json.Unmarshal(key, &redisMessage)
	return redisMessage
}
