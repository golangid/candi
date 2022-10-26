package taskqueueworker

import (
	"time"

	"github.com/golangid/candi/candiutils"
)

type (
	option struct {
		queue                    QueueStorage
		persistent               Persistent
		secondaryPersistent      Persistent
		maxClientSubscriber      int
		autoRemoveClientInterval time.Duration
		dashboardBanner          string
		dashboardPort            uint16
		debugMode                bool
		locker                   candiutils.Locker
	}

	// OptionFunc type
	OptionFunc func(*option)
)

// SetQueue option func
func SetQueue(q QueueStorage) OptionFunc {
	return func(o *option) {
		o.queue = q
	}
}

// SetPersistent option func
func SetPersistent(p Persistent) OptionFunc {
	return func(o *option) {
		o.persistent = p
	}
}

// SetSecondaryPersistent option func
func SetSecondaryPersistent(p Persistent) OptionFunc {
	return func(o *option) {
		o.secondaryPersistent = p
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

// SetExternalWorkerHost option func, setting worker host for add job, if not empty default using http request when add job
func SetExternalWorkerHost(host string) OptionFunc {
	externalWorkerHost = host
	return func(o *option) {
		externalWorkerHost = host
	}
}
