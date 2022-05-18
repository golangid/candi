package taskqueueworker

import (
	"time"

	"github.com/golangid/candi/candiutils"
)

type (
	option struct {
		tracingDashboard         string
		maxClientSubscriber      int
		autoRemoveClientInterval time.Duration
		dashboardBanner          string
		dashboardPort            uint16
		debugMode                bool
		locker                   candiutils.Locker
		maxConcurrentAddJob      int
		maxConcurrentBroadcast   int
	}

	// OptionFunc type
	OptionFunc func(*option)
)

// SetTracingDashboard option func
func SetTracingDashboard(host string) OptionFunc {
	return func(o *option) {
		o.tracingDashboard = host
	}
}

// SetMaxClientSubscriber option func
func SetMaxClientSubscriber(max int) OptionFunc {
	return func(o *option) {
		o.maxClientSubscriber = max
	}
}

// SetAutoRemoveClientInterval option func
func SetAutoRemoveClientInterval(d time.Duration) OptionFunc {
	return func(o *option) {
		o.autoRemoveClientInterval = d
	}
}

// SetDashboardBanner option func
func SetDashboardBanner(banner string) OptionFunc {
	return func(o *option) {
		o.dashboardBanner = banner
	}
}

// SetDashboardHTTPPort option func
func SetDashboardHTTPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.dashboardPort = port
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetLocker option func
func SetLocker(locker candiutils.Locker) OptionFunc {
	return func(o *option) {
		o.locker = locker
	}
}

// SetMaxConcurrentAddJob option func
func SetMaxConcurrentAddJob(max int) OptionFunc {
	return func(o *option) {
		o.maxConcurrentAddJob = max
	}
}

// SetMaxConcurrentBroadcast option func
func SetMaxConcurrentBroadcast(max int) OptionFunc {
	return func(o *option) {
		o.maxConcurrentBroadcast = max
	}
}
