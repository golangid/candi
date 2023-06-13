package tracer

import (
	"context"
	"runtime"
	"sync"

	"github.com/golangid/candi/candishared"
)

var (
	once         sync.Once
	activeTracer PlatformType = &noopTracer{}
)

// Tracer for trace
type Tracer interface {
	Context() context.Context
	NewContext() context.Context
	Tags() map[string]interface{}
	SetTag(key string, value interface{})
	InjectRequestHeader(header map[string]string)
	SetError(err error)
	Log(key string, value interface{})
	Finish(opts ...FinishOptionFunc)
}

// PlatformType define tracing platform. example using jaeger, sentry, aws x-ray, etc
type PlatformType interface {
	StartSpan(ctx context.Context, opName string) Tracer
	StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer
	GetTraceID(ctx context.Context) string
	GetTraceURL(ctx context.Context) string
}

// SetTracerPlatformType function for set tracer platform
func SetTracerPlatformType(t PlatformType) {
	once.Do(func() { activeTracer = t })
}

// StartTrace starting trace child span from parent span
func StartTrace(ctx context.Context, operationName string) Tracer {
	if candishared.GetValueFromContext(ctx, skipTracer) != nil {
		return &noopTracer{ctx}
	}

	return activeTracer.StartSpan(ctx, operationName)
}

// StartTraceWithContext starting trace child span from parent span, returning tracer and context
func StartTraceWithContext(ctx context.Context, operationName string) (Tracer, context.Context) {
	t := StartTrace(ctx, operationName)
	return t, t.Context()
}

// StartTraceFromHeader starting trace from root app handler based on header
func StartTraceFromHeader(ctx context.Context, operationName string, header map[string]string) (Tracer, context.Context) {
	if candishared.GetValueFromContext(ctx, skipTracer) != nil {
		return &noopTracer{ctx}, ctx
	}

	tc := activeTracer.StartRootSpan(ctx, operationName, header)
	return tc, tc.Context()
}

// GetTraceID func
func GetTraceID(ctx context.Context) string {
	return activeTracer.GetTraceID(ctx)
}

// GetTraceURL log trace url
func GetTraceURL(ctx context.Context) (u string) {
	return activeTracer.GetTraceURL(ctx)
}

// LogStackTrace log stack trace in recover panic
func LogStackTrace(trace Tracer) {
	const size = 2 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	trace.Log("stacktrace_detail", buf)
}

type noopTracer struct{ ctx context.Context }

func (n noopTracer) Context() context.Context                   { return n.ctx }
func (n noopTracer) NewContext() context.Context                { return n.ctx }
func (noopTracer) Tags() map[string]interface{}                 { return map[string]interface{}{} }
func (noopTracer) SetTag(key string, value interface{})         { return }
func (noopTracer) InjectRequestHeader(header map[string]string) { return }
func (noopTracer) SetError(err error)                           { return }
func (noopTracer) Log(key string, value interface{})            { return }
func (noopTracer) Finish(opts ...FinishOptionFunc)              { return }
func (noopTracer) GetTraceID(ctx context.Context) (u string)    { return }
func (noopTracer) GetTraceURL(ctx context.Context) (u string)   { return }
func (n noopTracer) StartSpan(ctx context.Context, opName string) Tracer {
	n.ctx = ctx
	return &n
}
func (n noopTracer) StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer {
	n.ctx = ctx
	return &n
}

var skipTracer candishared.ContextKey = "nooptracer"

// SkipTraceContext inject to context for skip span tracer
func SkipTraceContext(ctx context.Context) context.Context {
	return candishared.SetToContext(ctx, skipTracer, struct{}{})
}
