package tracer

import "context"

type (
	// Option for init tracer option
	Option struct {
		agentHost        string
		traceDashboard   string
		level            string
		buildNumberTag   string
		maxGoroutineTag  int
		logAllSpan       bool
		errorWhitelist   []error
		traceIDExtractor func(context.Context) string
	}

	// OptionFunc func
	OptionFunc func(*Option)

	// FinishOption for option when trace is finished
	FinishOption struct {
		Tags                 map[string]any
		Err                  error
		WithStackTraceDetail bool
		RecoverFunc          func(panicMessage any)
		OnFinish             func()
	}

	// FinishOptionFunc func
	FinishOptionFunc func(*FinishOption)
)

// OptionSetAgentHost option func
func OptionSetAgentHost(agent string) OptionFunc {
	return func(o *Option) {
		o.agentHost = agent
	}
}

// OptionSetLevel option func
func OptionSetLevel(level string) OptionFunc {
	return func(o *Option) {
		o.level = level
	}
}

// OptionSetBuildNumberTag option func
func OptionSetBuildNumberTag(number string) OptionFunc {
	return func(o *Option) {
		o.buildNumberTag = number
	}
}

// OptionSetMaxGoroutineTag option func
func OptionSetMaxGoroutineTag(max int) OptionFunc {
	return func(o *Option) {
		o.maxGoroutineTag = max
	}
}

// OptionSetTraceDashboardURL option func
func OptionSetTraceDashboardURL(dashboardURL string) OptionFunc {
	return func(o *Option) {
		o.traceDashboard = dashboardURL
	}
}

// OptionSetErrorWhitelist option func, set no error if error in whitelist
func OptionSetErrorWhitelist(errs []error) OptionFunc {
	return func(o *Option) {
		o.errorWhitelist = errs
	}
}

// OptionAddErrorWhitelist option func, add no error if error in whitelist
func OptionAddErrorWhitelist(errs ...error) OptionFunc {
	return func(o *Option) {
		o.errorWhitelist = append(o.errorWhitelist, errs...)
	}
}

// OptionSetTraceIDExtractor option func, set trace id extractor
func OptionSetTraceIDExtractor(extractor func(context.Context) string) OptionFunc {
	return func(o *Option) {
		o.traceIDExtractor = extractor
	}
}

// OptionSetLogAllSpan option func
func OptionSetLogAllSpan() OptionFunc {
	return func(o *Option) {
		o.logAllSpan = true
	}
}

// FinishWithError option for add error when finish
func FinishWithError(err error) FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.Err = err
	}
}

// FinishWithAdditionalTags option for add tag when finish
func FinishWithAdditionalTags(tags map[string]any) FinishOptionFunc {
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

// FinishWithRecoverPanic option for add recover func if panic
func FinishWithRecoverPanic(recoverFunc func(panicMessage any)) FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.RecoverFunc = recoverFunc
	}
}

// FinishWithFunc option for add callback function before finish span
func FinishWithFunc(finishFunc func()) FinishOptionFunc {
	return func(fo *FinishOption) {
		fo.OnFinish = finishFunc
	}
}
