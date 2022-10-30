/*
 * Source: github.com/gorhill/cronexpr with customization
 * Move to this package because repository in github has been deprecated
 */

package cronexpr

import (
	"fmt"
	"sort"
	"time"
)

// A Expression represents a specific cron time expression as defined at
// <https://github.com/gorhill/cronexpr#implementation>
type expression struct {
	expression             string
	secondList             []int
	minuteList             []int
	hourList               []int
	daysOfMonth            map[int]bool
	workdaysOfMonth        map[int]bool
	lastDayOfMonth         bool
	lastWorkdayOfMonth     bool
	daysOfMonthRestricted  bool
	actualDaysOfMonthList  []int
	monthList              []int
	daysOfWeek             map[int]bool
	specificWeekDaysOfWeek map[int]bool
	lastWeekDaysOfWeek     map[int]bool
	daysOfWeekRestricted   bool
	yearList               []int
}

// MustParse returns a new Expression pointer. It expects a well-formed cron
// expression. If a malformed cron expression is supplied, it will `panic`.
// See <https://github.com/gorhill/cronexpr#implementation> for documentation
// about what is a well-formed cron expression from this library's point of
// view.
func MustParse(cronLine string) Schedule {
	expr, err := Parse(cronLine)
	if err != nil {
		panic(err)
	}
	return expr
}

// Parse returns a new Expression pointer. An error is returned if a malformed
// cron expression is supplied.
// See <https://github.com/gorhill/cronexpr#implementation> for documentation
// about what is a well-formed cron expression from this library's point of
// view.
func Parse(cronLine string) (Schedule, error) {

	// Maybe one of the built-in aliases is being used
	cron := cronNormalizer.Replace(cronLine)

	indices := fieldFinder.FindAllStringIndex(cron, -1)
	fieldCount := len(indices)
	if fieldCount < 5 {
		return nil, fmt.Errorf("invalid cron expression")
	}
	// ignore fields beyond 7th
	if fieldCount > 7 {
		fieldCount = 7
	}

	var expr = expression{}
	var field = 0
	var err error

	// second field (optional)
	if fieldCount == 7 {
		err = expr.secondFieldHandler(cron[indices[field][0]:indices[field][1]])
		if err != nil {
			return nil, err
		}
		field += 1
	} else {
		expr.secondList = []int{0}
	}

	// minute field
	err = expr.minuteFieldHandler(cron[indices[field][0]:indices[field][1]])
	if err != nil {
		return nil, err
	}
	field += 1

	// hour field
	err = expr.hourFieldHandler(cron[indices[field][0]:indices[field][1]])
	if err != nil {
		return nil, err
	}
	field += 1

	// day of month field
	err = expr.domFieldHandler(cron[indices[field][0]:indices[field][1]])
	if err != nil {
		return nil, err
	}
	field += 1

	// month field
	err = expr.monthFieldHandler(cron[indices[field][0]:indices[field][1]])
	if err != nil {
		return nil, err
	}
	field += 1

	// day of week field
	err = expr.dowFieldHandler(cron[indices[field][0]:indices[field][1]])
	if err != nil {
		return nil, err
	}
	field += 1

	// year field
	if field < fieldCount {
		err = expr.yearFieldHandler(cron[indices[field][0]:indices[field][1]])
		if err != nil {
			return nil, err
		}
	} else {
		expr.yearList = yearDescriptor.defaultList
	}

	return &expr, nil
}

// Next returns the closest time instant immediately following `fromTime` which
// matches the cron expression `expr`.
//
// The `time.Location` of the returned time instant is the same as that of
// `fromTime`.
//
// The zero value of time.Time is returned if no matching time instant exists
// or if a `fromTime` is itself a zero value.
func (expr *expression) Next(fromTime time.Time) time.Time {
	// Special case
	if fromTime.IsZero() {
		return fromTime
	}

	// Since expr.nextSecond()-expr.nextMonth() expects that the
	// supplied time stamp is a perfect match to the underlying cron
	// expression, and since this function is an entry point where `fromTime`
	// does not necessarily matches the underlying cron expression,
	// we first need to ensure supplied time stamp matches
	// the cron expression. If not, this means the supplied time
	// stamp falls in between matching time stamps, thus we move
	// to closest future matching immediately upon encountering a mismatching
	// time stamp.

	// year
	v := fromTime.Year()
	i := sort.SearchInts(expr.yearList, v)
	if i == len(expr.yearList) {
		return time.Time{}
	}
	if v != expr.yearList[i] {
		return expr.nextYear(fromTime)
	}
	// month
	v = int(fromTime.Month())
	i = sort.SearchInts(expr.monthList, v)
	if i == len(expr.monthList) {
		return expr.nextYear(fromTime)
	}
	if v != expr.monthList[i] {
		return expr.nextMonth(fromTime)
	}

	expr.actualDaysOfMonthList = expr.calculateActualDaysOfMonth(fromTime.Year(), int(fromTime.Month()))
	if len(expr.actualDaysOfMonthList) == 0 {
		return expr.nextMonth(fromTime)
	}

	// day of month
	v = fromTime.Day()
	i = sort.SearchInts(expr.actualDaysOfMonthList, v)
	if i == len(expr.actualDaysOfMonthList) {
		return expr.nextMonth(fromTime)
	}
	if v != expr.actualDaysOfMonthList[i] {
		return expr.nextDayOfMonth(fromTime)
	}
	// hour
	v = fromTime.Hour()
	i = sort.SearchInts(expr.hourList, v)
	if i == len(expr.hourList) {
		return expr.nextDayOfMonth(fromTime)
	}
	if v != expr.hourList[i] {
		return expr.nextHour(fromTime)
	}
	// minute
	v = fromTime.Minute()
	i = sort.SearchInts(expr.minuteList, v)
	if i == len(expr.minuteList) {
		return expr.nextHour(fromTime)
	}
	if v != expr.minuteList[i] {
		return expr.nextMinute(fromTime)
	}
	// second
	v = fromTime.Second()
	i = sort.SearchInts(expr.secondList, v)
	if i == len(expr.secondList) {
		return expr.nextMinute(fromTime)
	}

	// If we reach this point, there is nothing better to do
	// than to move to the next second

	return expr.nextSecond(fromTime)
}

// NextInterval method
func (expr *expression) NextInterval(fromTime time.Time) time.Duration {
	return expr.Next(fromTime).Sub(fromTime)
}
