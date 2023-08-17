package tracer

type (
	// Option for init tracer option
	Option struct {
		AgentHost       string
		TraceDashboard  string
		Level           string
		BuildNumberTag  string
		MaxGoroutineTag int
	}

	// OptionFunc func
	OptionFunc func(*Option)

	// FinishOption for option when trace is finished
	FinishOption struct {
		Tags                 map[string]interface{}
		Error                error
		WithStackTraceDetail bool
	}

	// FinishOptionFunc func
	FinishOptionFunc func(*FinishOption)
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

// OptionSetTraceDashboardURL option func
func OptionSetTraceDashboardURL(dashboardURL string) OptionFunc {
	return func(o *Option) {
		o.TraceDashboard = dashboardURL
	}
}

// FinishWithError option for add error when finish
func FinishWithError(err error) FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.Error = err
	}
}

// FinishWithAdditionalTags option for add tag when finish
func FinishWithAdditionalTags(tags map[string]interface{}) FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.Tags = tags
	}
}

// FinishWithStackTraceDetail option for add stack trace detail
func FinishWithStackTraceDetail() FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.WithStackTraceDetail = true
	}
}
