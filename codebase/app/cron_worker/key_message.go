package cronworker

import "encoding/json"

const (
	lockPattern = "%s:cron-worker-lock:%s"
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

// CreateCronJobKey helper
/*
Allowed interval:

* cron expression, example: * * * * *

* standard time duration string, example: 2s, 10m

* custom start time and repeat duration, example:
	- 23:00@daily, will repeated at 23:00 every day
	- 23:00@weekly, will repeated at 23:00 every week
	- 23:00@10s, will repeated at 23:00 and next repeat every 10 seconds
*/
func CreateCronJobKey(jobName, args, interval string) string {
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
