package taskqueueworker

import "time"

// ErrorRetrier .
type ErrorRetrier struct {
	Delay   time.Duration
	Message string
}

// Error implement error
func (e *ErrorRetrier) Error() string {
	return e.Message
}
