package tracer

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi"
	"github.com/golangid/candi/config/env"
	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	config "github.com/uber/jaeger-client-go/config"
)

// InitOpenTracing init jaeger tracing
func InitOpenTracing(serviceName string, opts ...OptionFunc) error {
	option := Option{
		AgentHost:       env.BaseEnv().JaegerTracingHost,
		Level:           env.BaseEnv().Environment,
		BuildNumberTag:  env.BaseEnv().BuildNumber,
		MaxGoroutineTag: env.BaseEnv().MaxGoroutines,
	}
	urlAgent, err := url.Parse("//" + env.BaseEnv().JaegerTracingHost)
	if urlAgent != nil && err == nil {
		option.TraceDashboard = fmt.Sprintf("http://%s:16686/trace", urlAgent.Hostname())
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
	SetTracerPlatformType(&jaegerPlatform{
		dashboardURL: option.TraceDashboard,
	})
	return nil
}

type jaegerPlatform struct {
	dashboardURL string
}

func (j *jaegerPlatform) StartSpan(ctx context.Context, operationName string) Tracer {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		// init new span
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName)
	} else {
		span = opentracing.GlobalTracer().StartSpan(operationName, opentracing.ChildOf(span.Context()))
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	_, callerFile, callerLine, _ := runtime.Caller(4)
	span.LogKV("caller", callerFile+":"+strconv.Itoa(callerLine))
	return &jaegerTraceImpl{
		ctx:  ctx,
		span: span,
	}
}
func (j *jaegerPlatform) StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer {

	if header == nil {
		header = map[string]string{}
	}

	var span opentracing.Span
	globalTracer := opentracing.GlobalTracer()
	if spanCtx, err := globalTracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(header)); err != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName)
		ext.SpanKindRPCServer.Set(span)
	} else {
		span = globalTracer.StartSpan(operationName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	return &jaegerTraceImpl{
		ctx:  ctx,
		span: span,
	}
}
func (j *jaegerPlatform) GetTraceID(ctx context.Context) string {
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
func (j *jaegerPlatform) GetTraceURL(ctx context.Context) (u string) {
	if ctx == nil {
		return j.dashboardURL
	}
	traceID := j.GetTraceID(ctx)
	if traceID == "" {
		return "<disabled>"
	}

	return fmt.Sprintf("%s/%s", j.dashboardURL, traceID)
}

type jaegerTraceImpl struct {
	ctx  context.Context
	span opentracing.Span
	tags map[string]interface{}
}

// Context get active context
func (t *jaegerTraceImpl) Context() context.Context {
	return t.ctx
}

// Tags create tags in tracer span
func (t *jaegerTraceImpl) Tags() map[string]interface{} {
	if t.tags == nil {
		t.tags = make(map[string]interface{})
	}
	return t.tags
}

// SetTag set tags in tracer span
func (t *jaegerTraceImpl) SetTag(key string, value interface{}) {
	if t.span == nil {
		return
	}

	if t.tags == nil {
		t.tags = make(map[string]interface{})
	}
	t.tags[key] = value
}

// InjectRequestHeader to continue tracer with custom header carrier
func (t *jaegerTraceImpl) InjectRequestHeader(header map[string]string) {
	if t.span == nil {
		return
	}

	ext.SpanKindRPCClient.Set(t.span)
	t.span.Tracer().Inject(
		t.span.Context(),
		opentracing.HTTPHeaders,
		opentracing.TextMapCarrier(header),
	)
}

// NewContext to continue tracer with new context
func (t *jaegerTraceImpl) NewContext() context.Context {
	return opentracing.ContextWithSpan(context.Background(), t.span)
}

// SetError set error in span
func (t *jaegerTraceImpl) SetError(err error) {
	if t.span == nil || err == nil {
		return
	}

	ext.Error.Set(t.span, true)
	t.span.SetTag("error.message", err.Error())

	buff := make([]byte, 1<<10)
	buff = buff[:runtime.Stack(buff, false)]
	t.span.LogKV("stacktrace", toValue(buff))
}

// SetError log data
func (t *jaegerTraceImpl) Log(key string, value interface{}) {
	Log(t.ctx, key, value)
}

// Finish trace with additional tags data, must in deferred function
func (t *jaegerTraceImpl) Finish(opts ...FinishOptionFunc) {
	if t.span == nil {
		return
	}

	var finishOpt FinishOption
	for _, opt := range opts {
		if opt != nil {
			opt(&finishOpt)
		}
	}

	if finishOpt.Tags != nil && t.tags == nil {
		t.tags = make(map[string]interface{})
	}

	for k, v := range finishOpt.Tags {
		t.tags[k] = v
	}

	for k, v := range t.tags {
		t.span.SetTag(k, toValue(v))
	}

	if finishOpt.Error != nil {
		t.SetError(finishOpt.Error)
	} else if finishOpt.WithStackTraceDetail {
		buff := make([]byte, 1<<10)
		buff = buff[:runtime.Stack(buff, false)]
		t.span.LogKV("stacktrace", toValue(buff))
	}

	t.span.Finish()
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
