package tracer

import (
	"context"
	"runtime"
	"sync"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/interfaces"
)

var (
	once         sync.Once
	activeTracer PlatformType = &NoopTracer{}
)

// Tracer for trace
type Tracer interface {
	Context() context.Context
	NewContext() context.Context
	SetTag(key string, value any)
	InjectRequestHeader(header map[string]string)
	SetError(err error)
	Log(key string, value any)
	Finish(opts ...FinishOptionFunc)
}

// PlatformType define tracing platform. example using jaeger, sentry, aws x-ray, etc
type PlatformType interface {
	StartSpan(ctx context.Context, opName string) Tracer
	StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer
	GetTraceID(ctx context.Context) string
	GetTraceURL(ctx context.Context) string
	interfaces.Closer
}

// SetTracerPlatformType function for set tracer platform
func SetTracerPlatformType(t PlatformType) {
	once.Do(func() { activeTracer = t })
}

// IsTracerActive check tracer has been initialized with platform
func IsTracerActive() bool {
	_, ok := activeTracer.(*NoopTracer)
	return !ok
}

// StartTrace starting trace child span from parent span
func StartTrace(ctx context.Context, operationName string) Tracer {
	if candishared.GetValueFromContext(ctx, skipTracer) != nil {
		return &NoopTracer{ctx}
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
		return &NoopTracer{ctx}, ctx
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

type NoopTracer struct{ ctx context.Context }

func (n NoopTracer) Context() context.Context                   { return n.ctx }
func (n NoopTracer) NewContext() context.Context                { return n.ctx }
func (NoopTracer) SetTag(key string, value any)                 {}
func (NoopTracer) InjectRequestHeader(header map[string]string) {}
func (NoopTracer) SetError(err error)                           {}
func (NoopTracer) Log(key string, value any)                    {}
func (NoopTracer) Finish(opts ...FinishOptionFunc) {
	var finishOpt FinishOption
	for _, opt := range opts {
		opt(&finishOpt)
	}
	if finishOpt.RecoverFunc != nil {
		if rec := recover(); rec != nil {
			finishOpt.RecoverFunc(rec)
		}
	}
	if finishOpt.OnFinish != nil {
		finishOpt.OnFinish()
	}
}
func (NoopTracer) GetTraceID(ctx context.Context) (u string)  { return }
func (NoopTracer) GetTraceURL(ctx context.Context) (u string) { return }
func (NoopTracer) Disconnect(ctx context.Context) error       { return nil }
func (n NoopTracer) StartSpan(ctx context.Context, opName string) Tracer {
	n.ctx = ctx
	return &n
}
func (n NoopTracer) StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer {
	n.ctx = ctx
	return &n
}

var skipTracer candishared.ContextKey = "nooptracer"

// SkipTraceContext inject to context for skip span tracer
func SkipTraceContext(ctx context.Context) context.Context {
	return candishared.SetToContext(ctx, skipTracer, struct{}{})
}
