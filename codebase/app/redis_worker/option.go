package redisworker

import "github.com/golangid/candi/codebase/interfaces"

type (
	option struct {
		maxGoroutines int
		locker        interfaces.Locker
		debugMode     bool
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		maxGoroutines: 10,
		debugMode:     true,
	}
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
