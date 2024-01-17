package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// WithTracerFunc functional with Tracer instance in function params
func WithTracerFunc(ctx context.Context, operationName string, fn func(context.Context, Tracer)) {
	t, ctx := StartTraceWithContext(ctx, operationName)
	defer t.Finish()

	fn(ctx, t)
}

func toValue(v any) (res any) {
	var str string
	switch val := v.(type) {
	case error:
		if val != nil {
			str = val.Error()
		}
	case string:
		str = val
	case bool, int8, int16, int32, int, int64, float32, float64:
		return v
	case []byte:
		str = candihelper.ByteToString(val)
	default:
		b, _ := json.Marshal(val)
		str = candihelper.ByteToString(b)
	}

	if len(str) >= int(env.BaseEnv().JaegerMaxPacketSize) {
		return fmt.Sprintf("<<Overflow, cannot show data. Size is = %d bytes, JAEGER_MAX_PACKET_SIZE = %d bytes>>",
			len(str),
			env.BaseEnv().JaegerMaxPacketSize)
	}

	return logger.MaskLog(str)
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

// Log trace
func Log(ctx context.Context, key string, value interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.LogKV(key, toValue(value))
}

// LogEvent trace
func LogEvent(ctx context.Context, event string, payload ...interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if payload != nil {
		for _, p := range payload {
			if e, ok := p.(error); ok && e != nil {
				ext.Error.Set(span, true)
			}
			span.LogKV(event, toValue(p))
		}
	} else {
		span.LogKV(event)
	}
}

func parseCaller(pc uintptr, file string, line int, ok bool) (caller string) {
	if !ok {
		return
	}

	if strings.HasSuffix(file, "candi/tracer/jaeger.go") {
		return
	}
	return file + ":" + strconv.Itoa(line)
}
