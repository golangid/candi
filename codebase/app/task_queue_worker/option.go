package taskqueueworker

import "time"

type (
	option struct {
		JaegerTracingDashboard   string
		MaxClientSubscriber      int
		AutoRemoveClientInterval time.Duration
	}

	// OptionFunc type
	OptionFunc func(*option)
)

// SetJaegerTracingDashboard option func
func SetJaegerTracingDashboard(host string) OptionFunc {
	return func(o *option) {
		o.JaegerTracingDashboard = host
	}
}

// SetMaxClientSubscriber option func
func SetMaxClientSubscriber(max int) OptionFunc {
	return func(o *option) {
		o.MaxClientSubscriber = max
	}
}

// SetAutoRemoveClientInterval option func
func SetAutoRemoveClientInterval(d time.Duration) OptionFunc {
	return func(o *option) {
		o.AutoRemoveClientInterval = d
	}
}
