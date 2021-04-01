package candihelper

import (
	"encoding/json"
)

// CronJobKey model
type CronJobKey struct {
	JobName  string `json:"jobName"`
	Args     string `json:"args"`
	Interval string `json:"interval"`
}

// String implement stringer
func (c CronJobKey) String() string {
	b, _ := json.Marshal(c)
	return string(b)
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
func CronJobKeyToString(jobName, args, interval string) string {
	return CronJobKey{
		JobName: jobName, Args: args, Interval: interval,
	}.String()
}

// ParseCronJobKey helper
func ParseCronJobKey(str string) (jobName, args, interval string) {
	var cronKey CronJobKey
	json.Unmarshal([]byte(str), &cronKey)
	return cronKey.JobName, cronKey.Args, cronKey.Interval
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
