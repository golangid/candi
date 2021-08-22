package cronworker

import "pkg.agungdp.dev/candi/candiutils"

type (
	option struct {
		maxGoroutines int
		consul        *candiutils.Consul
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

// SetConsul option func
func SetConsul(consul *candiutils.Consul) OptionFunc {
	return func(o *option) {
		o.consul = consul
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}
