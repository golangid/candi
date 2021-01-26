package candihelper

import (
	"encoding/json"
	"fmt"
	"strings"
)

type jobKey struct {
	JobName  string `json:"jobName"`
	Interval string `json:"interval,omitempty"`
	MaxRetry int    `json:"maxRetry,omitempty"`
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

// BuildRedisPubSubKeyTopic helper
func BuildRedisPubSubKeyTopic(handlerName string, payload interface{}) string {
	return fmt.Sprintf("%s~%s", strings.Replace(handlerName, "~", "", -1), ToBytes(payload))
}

// ParseRedisPubSubKeyTopic helper
func ParseRedisPubSubKeyTopic(str string) (handlerName, messageData string) {
	defer func() { recover() }()

	split := strings.Split(str, "~")
	handlerName = split[0]
	messageData = strings.Join(split[1:], "~")
	return
}
