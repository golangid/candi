package tracer

import (
	"context"
	"encoding/json"
	"runtime"
	"strconv"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WithTracerFunc functional with Tracer instance in function params
func WithTracerFunc(ctx context.Context, operationName string, fn func(context.Context, Tracer)) {
	t, ctx := StartTraceWithContext(ctx, operationName)
	defer t.Finish()

	fn(ctx, t)
}

func toOtelValue(v any) (res attribute.Value) {
	var str string
	switch val := v.(type) {
	case error:
		if val != nil {
			str = val.Error()
		}
	case string:
		str = val

	case bool:
		return attribute.BoolValue(val)
	case int8:
		return attribute.IntValue(int(val))
	case int16:
		return attribute.IntValue(int(val))
	case int32:
		return attribute.IntValue(int(val))
	case int:
		return attribute.IntValue(val)
	case int64:
		return attribute.Int64Value(val)
	case float32:
		return attribute.Float64Value(float64(val))
	case float64:
		return attribute.Float64Value(val)

	case []byte:
		str = candihelper.ByteToString(val)
	default:
		b, _ := json.Marshal(val)
		str = candihelper.ByteToString(b)
	}

	return attribute.StringValue(logger.MaskLog(str))
}

// SetError global func
// TODO: separate in each tracer platform
func SetError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span == nil || err == nil {
		return
	}
	span.SetStatus(codes.Error, err.Error())

	stackTrace := make([]byte, 1024)
	for {
		n := runtime.Stack(stackTrace, false)
		if n < len(stackTrace) {
			stackTrace = stackTrace[:n]
			break
		}
		stackTrace = make([]byte, 2*len(stackTrace))
	}
	span.AddEvent("", trace.WithAttributes(
		attribute.KeyValue{
			Key: attribute.Key("stacktrace"), Value: toOtelValue(stackTrace),
		},
	))
}

// Log trace
func Log(ctx context.Context, key string, value any) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.AddEvent("", trace.WithAttributes(
		attribute.KeyValue{
			Key: attribute.Key(key), Value: toOtelValue(value),
		},
	))
}

// LogEvent trace
func LogEvent(ctx context.Context, event string, payload ...any) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if payload != nil {
		for _, p := range payload {
			if e, ok := p.(error); ok && e != nil {
				span.SetStatus(codes.Error, e.Error())
			}
			span.AddEvent("", trace.WithAttributes(
				attribute.KeyValue{
					Key: attribute.Key(event), Value: toOtelValue(p),
				},
			))
		}
	} else {
		span.AddEvent(event)
	}
}

func parseCaller(_ uintptr, file string, line int, ok bool) (caller string) {
	if !ok {
		return
	}

	if strings.HasSuffix(file, "candi/tracer/jaeger.go") {
		return
	}
	return file + ":" + strconv.Itoa(line)
}
