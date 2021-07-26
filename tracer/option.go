package tracer

// Option for init tracer option
type (
	Option struct {
		AgentHost       string
		Level           string
		BuildNumberTag  string
		MaxGoroutineTag int
	}

	// OptionFunc func
	OptionFunc func(*Option)
)

// OptionSetAgentHost option func
func OptionSetAgentHost(agent string) OptionFunc {
	return func(o *Option) {
		o.AgentHost = agent
	}
}

// OptionSetLevel option func
func OptionSetLevel(level string) OptionFunc {
	return func(o *Option) {
		o.Level = level
	}
}

// OptionSetBuildNumberTag option func
func OptionSetBuildNumberTag(number string) OptionFunc {
	return func(o *Option) {
		o.BuildNumberTag = number
	}
}

// OptionSetMaxGoroutineTag option func
func OptionSetMaxGoroutineTag(max int) OptionFunc {
	return func(o *Option) {
		o.MaxGoroutineTag = max
	}
}
