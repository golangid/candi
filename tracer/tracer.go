package tracer

import (
	"context"
	"net/http"
	"sync"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/interfaces"
	"google.golang.org/grpc/metadata"
)

var (
	once         sync.Once
	activeTracer PlatformType = &noopTracer{}
)

// PlatformType define tracing platform. example using jaeger, sentry, aws x-ray, etc
type PlatformType interface {
	StartSpan(ctx context.Context, opName string) interfaces.Tracer
	StartRootSpan(ctx context.Context, operationName string, header map[string]string) interfaces.Tracer
}

// SetTracerPlatformType function for set tracer platform
func SetTracerPlatformType(t PlatformType) {
	once.Do(func() { activeTracer = t })
}

// StartTrace starting trace child span from parent span
func StartTrace(ctx context.Context, operationName string) interfaces.Tracer {
	if candishared.GetValueFromContext(ctx, skipTracer) != nil {
		return &noopTracer{ctx}
	}

	return activeTracer.StartSpan(ctx, operationName)
}

// StartTraceWithContext starting trace child span from parent span, returning tracer and context
func StartTraceWithContext(ctx context.Context, operationName string) (interfaces.Tracer, context.Context) {
	t := StartTrace(ctx, operationName)
	return t, t.Context()
}

// StartTraceFromHeader starting trace from root app handler based on header
func StartTraceFromHeader(ctx context.Context, operationName string, header map[string]string) (interfaces.Tracer, context.Context) {

	tc := activeTracer.StartRootSpan(ctx, operationName, header)
	return tc, tc.Context()
}

type noopTracer struct{ ctx context.Context }

func (n noopTracer) Context() context.Context                      { return n.ctx }
func (noopTracer) Tags() map[string]interface{}                    { return map[string]interface{}{} }
func (noopTracer) SetTag(key string, value interface{})            { return }
func (noopTracer) InjectRequestHeader(header map[string]string)    { return }
func (noopTracer) SetError(err error)                              { return }
func (noopTracer) Log(key string, value interface{})               { return }
func (noopTracer) Finish(additionalTags ...map[string]interface{}) { return }

func (n noopTracer) StartSpan(ctx context.Context, opName string) interfaces.Tracer {
	n.ctx = ctx
	return &n
}
func (n noopTracer) StartRootSpan(ctx context.Context, operationName string, header map[string]string) interfaces.Tracer {
	n.ctx = ctx
	return &n
}

// Deprecated: use InjectRequestHeader
func (noopTracer) InjectHTTPHeader(req *http.Request) { return }

// Deprecated: use InjectRequestHeader
func (noopTracer) InjectGRPCMetadata(md metadata.MD) { return }
