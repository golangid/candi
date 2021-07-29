package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	config "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc/metadata"
	"pkg.agungdp.dev/candi"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
)

// InitOpenTracing init jaeger tracing
func InitOpenTracing(serviceName string, opts ...OptionFunc) error {
	option := Option{
		AgentHost:       env.BaseEnv().JaegerTracingHost,
		Level:           env.BaseEnv().Environment,
		BuildNumberTag:  env.BaseEnv().BuildNumber,
		MaxGoroutineTag: env.BaseEnv().MaxGoroutines,
	}

	for _, opt := range opts {
		opt(&option)
	}

	if option.Level != "" {
		serviceName = fmt.Sprintf("%s-%s", serviceName, strings.ToLower(option.Level))
	}
	defaultTags := []opentracing.Tag{
		{Key: "num_cpu", Value: runtime.NumCPU()},
		{Key: "go_version", Value: runtime.Version()},
		{Key: "candi_version", Value: candi.Version},
	}
	if option.MaxGoroutineTag != 0 {
		defaultTags = append(defaultTags, opentracing.Tag{
			Key: "max_goroutines", Value: option.MaxGoroutineTag,
		})
	}
	if option.BuildNumberTag != "" {
		defaultTags = append(defaultTags, opentracing.Tag{
			Key: "build_number", Value: option.BuildNumberTag,
		})
	}
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  option.AgentHost,
		},
		ServiceName: serviceName,
		Tags:        defaultTags,
	}
	tracer, _, err := cfg.NewTracer(config.MaxTagValueLength(math.MaxInt32))
	if err != nil {
		log.Printf("ERROR: cannot init opentracing connection: %v\n", err)
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

type jaegerImpl struct {
	ctx  context.Context
	span opentracing.Span
	tags map[string]interface{}
}

// StartTrace starting trace child span from parent span
func StartTrace(ctx context.Context, operationName string) interfaces.Tracer {
	if candishared.GetValueFromContext(ctx, skipTracer) != nil {
		return &jaegerImpl{ctx: ctx}
	}

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		// init new span
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName)
	} else {
		span = opentracing.GlobalTracer().StartSpan(operationName, opentracing.ChildOf(span.Context()))
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	return &jaegerImpl{
		ctx:  ctx,
		span: span,
	}
}

// StartTraceWithContext starting trace child span from parent span, returning tracer and context
func StartTraceWithContext(ctx context.Context, operationName string) (interfaces.Tracer, context.Context) {
	t := StartTrace(ctx, operationName)
	return t, t.Context()
}

// Context get active context
func (t *jaegerImpl) Context() context.Context {
	return t.ctx
}

// Tags create tags in tracer span
func (t *jaegerImpl) Tags() map[string]interface{} {
	t.tags = make(map[string]interface{})
	return t.tags
}

// SetTag set tags in tracer span
func (t *jaegerImpl) SetTag(key string, value interface{}) {
	if t.span == nil {
		return
	}

	if t.tags == nil {
		t.tags = make(map[string]interface{})
	}
	t.tags[key] = value
}

// InjectHTTPHeader to continue tracer to http request host
func (t *jaegerImpl) InjectHTTPHeader(req *http.Request) {
	if t.span == nil {
		return
	}
	ext.SpanKindRPCClient.Set(t.span)
	t.span.Tracer().Inject(
		t.span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

// InjectGRPCMetaData to continue tracer to grpc metadata context
func (t *jaegerImpl) InjectGRPCMetadata(md metadata.MD) {
	if t.span == nil {
		return
	}

	ext.SpanKindRPCClient.Set(t.span)
	t.span.Tracer().Inject(
		t.span.Context(),
		opentracing.HTTPHeaders,
		GRPCMetadataReaderWriter(md),
	)
}

// SetError set error in span
func (t *jaegerImpl) SetError(err error) {
	SetError(t.ctx, err)
}

// SetError log data
func (t *jaegerImpl) Log(key string, value interface{}) {
	Log(t.ctx, key, value)
}

// Finish trace with additional tags data, must in deferred function
func (t *jaegerImpl) Finish(additionalTags ...map[string]interface{}) {
	if t.span == nil {
		return
	}

	defer t.span.Finish()
	if additionalTags != nil && t.tags == nil {
		t.tags = make(map[string]interface{})
	}

	for _, tag := range additionalTags {
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
func Log(ctx context.Context, key string, value interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.LogKV(key, toString(value))
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

// WithTraceFuncTracer functional with Tracer instance in function params
func WithTraceFuncTracer(ctx context.Context, operationName string, fn func(t interfaces.Tracer)) {
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
	case []byte:
		s = string(val)
	default:
		b, _ := json.Marshal(val)
		s = string(b)
	}

	if len(s) >= int(env.BaseEnv().JaegerMaxPacketSize) {
		return fmt.Sprintf("<<Overflow, cannot show data. Size is = %d bytes, JAEGER_MAX_PACKET_SIZE = %d bytes>>", len(s), env.BaseEnv().JaegerMaxPacketSize)
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
	span.LogFields(otlog.String("stacktrace", string(debug.Stack())))
}

// GetTraceURL log trace url
func GetTraceURL(ctx context.Context) (u string) {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		return
	}

	urlAgent, err := url.Parse("//" + env.BaseEnv().JaegerTracingHost)
	if urlAgent != nil && err == nil {
		u = fmt.Sprintf("http://%s:16686/trace/%s", urlAgent.Hostname(), traceID)
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

var skipTracer candishared.ContextKey = "nooptracer"

// SkipTraceContext inject to context for skip span tracer
func SkipTraceContext(ctx context.Context) context.Context {
	return candishared.SetToContext(ctx, skipTracer, struct{}{})
}
