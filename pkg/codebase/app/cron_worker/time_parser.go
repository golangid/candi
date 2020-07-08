package cronworker

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	oneDay = 24 * time.Hour
)

// parseAtTime with input format HH:mm:ss, will repeat every day (24 hours) in the same time
func parseAtTime(t string) (duration, nextDuration time.Duration, err error) {
	ts := strings.Split(t, ":")
	if len(ts) < 2 || len(ts) > 3 {
		return 0, 0, errors.New("time format error")
	}

	var hour, min, sec int
	hour, err = strconv.Atoi(ts[0])
	if err != nil {
		return
	}

	min, err = strconv.Atoi(ts[1])
	if err != nil {
		return
	}

	if len(ts) == 3 {
		if sec, err = strconv.Atoi(ts[2]); err != nil {
			return
		}
	}

	if hour < 0 || hour > 23 || min < 0 || min > 59 || sec < 0 || sec > 59 {
		return 0, 0, errors.New("time format error")
	}

	now := time.Now()
	atTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, now.Location())
	if now.Before(atTime) {
		duration = atTime.Sub(now)
	} else {
		duration = oneDay - now.Sub(atTime)
	}
	nextDuration = oneDay

	return
}
