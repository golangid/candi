package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

// Tracer for trace
type Tracer interface {
	Context() context.Context
	Tags() map[string]interface{}
	InjectHTTPHeader(req *http.Request)
	InjectGRPCMetadata(md metadata.MD)
	SetError(err error)
	Finish(additionalTags ...map[string]interface{})
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
		// init new span
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

// InjectHTTPHeader to continue tracer to http request host
func (t *tracerImpl) InjectHTTPHeader(req *http.Request) {
	ext.SpanKindRPCClient.Set(t.span)
	t.span.Tracer().Inject(
		t.span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

// InjectGRPCMetaData to continue tracer to grpc metadata context
func (t *tracerImpl) InjectGRPCMetadata(md metadata.MD) {
	ext.SpanKindRPCClient.Set(t.span)
	t.span.Tracer().Inject(
		t.span.Context(),
		opentracing.HTTPHeaders,
		GRPCMetadataReaderWriter(md),
	)
}

// SetError set error in span
func (t *tracerImpl) SetError(err error) {
	SetError(t.ctx, err)
}

// Finish trace with additional tags data, must in deferred function
func (t *tracerImpl) Finish(tags ...map[string]interface{}) {
	defer t.span.Finish()

	if tags != nil && t.tags == nil {
		t.tags = make(map[string]interface{})
	}

	for _, tag := range tags {
		for k, v := range tag {
			t.tags[k] = v
		}
	}

	for k, v := range t.tags {
		t.span.SetTag(k, toString(v))
	}
	t.span.SetTag("num_goroutines", runtime.NumGoroutine())
}

// Log trace
func Log(ctx context.Context, event string, payload ...interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if payload != nil {
		for _, p := range payload {
			if e, ok := p.(error); ok && e != nil {
				ext.Error.Set(span, true)
			}
			span.LogEventWithPayload(event, toString(p))
		}
	} else {
		span.LogEvent(event)
	}
}

// WithTraceFunc functional with context and tags in function params
func WithTraceFunc(ctx context.Context, operationName string, fn func(context.Context, map[string]interface{})) {
	t := StartTrace(ctx, operationName)
	defer t.Finish()

	fn(t.Context(), t.Tags())
}

// WithTraceFuncTracer functional with Tracer in function params
func WithTraceFuncTracer(ctx context.Context, operationName string, fn func(t Tracer)) {
	t := StartTrace(ctx, operationName)
	defer t.Finish()

	fn(t)
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

	if len(s) >= maxPacketSize {
		s = fmt.Sprintf("overflow, size is: %d, max: %d", len(s), maxPacketSize)
	}
	return
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

// GetTraceURL log trace url
func GetTraceURL(ctx context.Context) (u string) {
	defer func() { recover() }()
	urlAgent, err := url.Parse("//" + env.BaseEnv().JaegerTracingHost)
	if err == nil {
		u = fmt.Sprintf("> tracing_url: %s:16686/trace/%s", urlAgent.Hostname(), GetTraceID(ctx))
	}
	return
}

// GRPCMetadataReaderWriter grpc metadata
type GRPCMetadataReaderWriter metadata.MD

// ForeachKey method
func (mrw GRPCMetadataReaderWriter) ForeachKey(handler func(string, string) error) error {
	for key, values := range mrw {
		for _, value := range values {
			dk, dv, err := metadata.DecodeKeyValue(key, value)
			if err != nil {
				return err
			}

			if err := handler(dk, dv); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set method
func (mrw GRPCMetadataReaderWriter) Set(key, value string) {
	// headers should be lowercase
	k := strings.ToLower(key)
	mrw[k] = append(mrw[k], value)
}
