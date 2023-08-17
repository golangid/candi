package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/golangid/candi/config/env"
	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// WithTraceFunc functional with context and tags in function params (DEPRECATED, use WithTracerFunc)
func WithTraceFunc(ctx context.Context, operationName string, fn func(context.Context, map[string]interface{})) {
	t := StartTrace(ctx, operationName)
	defer t.Finish()

	fn(t.Context(), t.Tags())
}

// WithTracerFunc functional with Tracer instance in function params
func WithTracerFunc(ctx context.Context, operationName string, fn func(context.Context, Tracer)) {
	t, ctx := StartTraceWithContext(ctx, operationName)
	defer t.Finish()

	fn(ctx, t)
}

func toValue(v interface{}) (s interface{}) {

	var str string
	switch val := v.(type) {

	case uint, uint64, int, int64, float32, float64, bool:
		return v

	case error:
		if val != nil {
			str = val.Error()
		}
	case string:
		str = val
	case []byte:
		str = string(val)
	default:
		b, _ := json.Marshal(val)
		str = string(b)
	}

	if len(str) >= int(env.BaseEnv().JaegerMaxPacketSize) {
		return fmt.Sprintf("<<Overflow, cannot show data. Size is = %d bytes, JAEGER_MAX_PACKET_SIZE = %d bytes>>",
			len(str),
			env.BaseEnv().JaegerMaxPacketSize)
	}

	return str
}

// SetError global func
// TODO: separate in each tracer platform
func SetError(ctx context.Context, err error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || err == nil {
		return
	}

	ext.Error.Set(span, true)
	span.SetTag("error.message", err.Error())

	stackTrace := make([]byte, 1024)
	for {
		n := runtime.Stack(stackTrace, false)
		if n < len(stackTrace) {
			stackTrace = stackTrace[:n]
			break
		}
		stackTrace = make([]byte, 2*len(stackTrace))
	}
	span.LogFields(otlog.String("stacktrace", string(stackTrace)))
}
