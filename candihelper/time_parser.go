package candihelper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// OneDay const
	OneDay = 24 * time.Hour
	// OneWeek const
	OneWeek = 7 * OneDay
	// OneMonth const
	OneMonth = 30 * OneDay
	// OneYear const
	OneYear = 12 * OneMonth

	// Daily const
	Daily = "daily"
	// Weekly const
	Weekly = "weekly"
	// Monthly const
	Monthly = "monthly"
	// Yearly const
	Yearly = "yearly"
)

// ParseDurationExpression with input format HH:mm:ss
func ParseDurationExpression(t string) (duration, nextDuration time.Duration, err error) {
	interval, err := time.ParseDuration(t)
	if err == nil {
		return interval, 0, nil
	}

	withDescriptors := strings.Split(t, "@")

	ts := strings.Split(withDescriptors[0], ":")
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

	// default value
	repeatDuration := OneDay

	if len(withDescriptors) > 1 {

		switch withDescriptors[1] {
		case Daily:
			repeatDuration = OneDay
		case Weekly:
			repeatDuration = OneWeek
		case Monthly:
			repeatDuration = OneMonth
		case Yearly:
			repeatDuration = OneYear
		default:
			repeatDuration, err = time.ParseDuration(withDescriptors[1])
			if err != nil {
				return 0, 0, fmt.Errorf(`invalid descriptor "%s" (must One of "daily", "weekly", "monthly", "yearly") or duration string`,
					withDescriptors[1])
			}
		}
	}

	now := time.Now()
	atTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, now.Location())
	if now.Before(atTime) {
		duration = atTime.Sub(now)
	} else {
		duration = OneDay - now.Sub(atTime)
	}

	if duration < 0 {
		duration *= -1
	}

	nextDuration = repeatDuration

	return
}
