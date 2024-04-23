package rabbitmqworker

type (
	option struct {
		consumerGroup string
		maxGoroutines int
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

// SetConsumerGroup option func, for log when startup
func SetConsumerGroup(consumerGroup string) OptionFunc {
	return func(o *option) {
		o.consumerGroup = consumerGroup
	}
}
