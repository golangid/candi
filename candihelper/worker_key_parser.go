package candihelper

import (
	"encoding/json"
)

type jobKey struct {
	JobName  string `json:"jobName"`
	Interval string `json:"interval,omitempty"`
}

// CronJobKeyToString helper
/*
Allowed interval:

* standard time duration string, example: 2s, 10m

* custom start time and repeat duration, example:
	- 23:00@daily, will repeated at 23:00 every day
	- 23:00@weekly, will repeated at 23:00 every week
	- 23:00@10s, will repeated at 23:00 and next repeat every 10 seconds
*/
func CronJobKeyToString(jobName string, interval string) string {
	b, _ := json.Marshal(jobKey{
		JobName: jobName, Interval: interval,
	})
	return string(b)
}

// ParseCronJobKey helper
func ParseCronJobKey(str string) (string, string) {
	var cronKey jobKey
	json.Unmarshal([]byte(str), &cronKey)
	return cronKey.JobName, cronKey.Interval
}

// RedisMessage model for redis subscriber key
type RedisMessage struct {
	HandlerName string `json:"h"`
	Message     string `json:"message"`
}

// BuildRedisPubSubKeyTopic helper
func BuildRedisPubSubKeyTopic(handlerName string, message interface{}) string {
	key, _ := json.Marshal(RedisMessage{HandlerName: handlerName, Message: string(ToBytes(message))})
	return string(key)
}

// ParseRedisPubSubKeyTopic helper
func ParseRedisPubSubKeyTopic(str string) (handlerName, messageData string) {
	var redisMessage RedisMessage
	json.Unmarshal([]byte(str), &redisMessage)
	return redisMessage.HandlerName, redisMessage.Message
}
