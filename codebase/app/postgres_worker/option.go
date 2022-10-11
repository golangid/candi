package postgresworker

import "github.com/golangid/candi/candiutils"

type (
	option struct {
		postgresDSN   string
		maxGoroutines int
		debugMode     bool
		locker        candiutils.Locker
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		maxGoroutines: 10,
		debugMode:     true,
		locker:        &candiutils.NoopLocker{},
	}
}

// SetPostgresDSN option func
func SetPostgresDSN(dsn string) OptionFunc {
	return func(o *option) {
		o.postgresDSN = dsn
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

// SetLocker option func
func SetLocker(locker candiutils.Locker) OptionFunc {
	return func(o *option) {
		o.locker = locker
	}
}
