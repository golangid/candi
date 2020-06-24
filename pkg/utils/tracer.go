package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Tracer for trace
type Tracer interface {
	Context() context.Context
	Tags() map[string]interface{}
	SetError(err error)
	Finish()
}

type tracerImpl struct {
	ctx  context.Context
	span opentracing.Span
	tags map[string]interface{}
}

// StartTrace starting trace child span from parent span
func StartTrace(ctx context.Context, operationName string) Tracer {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName)
	} else {
		span = opentracing.GlobalTracer().StartSpan(operationName, opentracing.ChildOf(span.Context()))
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	return &tracerImpl{
		ctx:  ctx,
		span: span,
	}
}

// Context get active context
func (t *tracerImpl) Context() context.Context {
	return t.ctx
}

// Tags create tags in tracer span
func (t *tracerImpl) Tags() map[string]interface{} {
	t.tags = make(map[string]interface{})
	return t.tags
}

// SetError set error in span
func (t *tracerImpl) SetError(err error) {
	SetError(t.ctx, err)
}

// Finish trace with additional tags data, must in deferred function
func (t *tracerImpl) Finish() {
	defer t.span.Finish()

	for k, v := range t.tags {
		t.span.SetTag(k, toString(v))
	}
}

func toString(v interface{}) (s string) {
	switch val := v.(type) {
	case error:
		if val != nil {
			s = val.Error()
		}
	case string:
		s = val
	case int:
		s = strconv.Itoa(val)
	default:
		b, _ := json.Marshal(val)
		s = string(b)
	}
	return
}

// WithTraceFunc functional with context and tags in function params
func WithTraceFunc(ctx context.Context, operationName string, fn func(context.Context, map[string]interface{})) {
	t := StartTrace(ctx, operationName)
	defer t.Finish()

	fn(t.Context(), t.Tags())
}

// GetTraceID func
func GetTraceID(ctx context.Context) string {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	traceID := fmt.Sprintf("%+v", span)
	splits := strings.Split(traceID, ":")
	if len(splits) > 0 {
		return splits[0]
	}

	return traceID
}

// SetError func
func SetError(ctx context.Context, err error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || err == nil {
		return
	}

	ext.Error.Set(span, true)
	span.SetTag("error.message", err.Error())
	span.SetTag("stacktrace", string(debug.Stack()))
}
