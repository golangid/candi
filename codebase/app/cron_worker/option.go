package cronworker

import (
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/interfaces"
)

type (
	option struct {
		maxGoroutines int
		debugMode     bool
		locker        interfaces.Locker
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption(service factory.ServiceFactory) option {
	opt := option{
		maxGoroutines: 10,
		debugMode:     true,
	}
	if redisPool := service.GetDependency().GetRedisPool(); redisPool != nil {
		opt.locker = candiutils.NewRedisLocker(redisPool.WritePool())
	} else {
		opt.locker = &candiutils.NoopLocker{}
	}
	return opt
}

// SetMaxGoroutines option func
func SetMaxGoroutines(maxGoroutines int) OptionFunc {
	return func(o *option) {
		o.maxGoroutines = maxGoroutines
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetLocker option func
func SetLocker(locker interfaces.Locker) OptionFunc {
	return func(o *option) {
		o.locker = locker
	}
}
