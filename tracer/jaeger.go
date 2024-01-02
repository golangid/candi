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

// InitJaeger init jaeger tracing
func InitJaeger(serviceName string, opts ...OptionFunc) error {
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
		log.Printf("ERROR: cannot init jaeger opentracing connection: %v\n", err)
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	SetTracerPlatformType(&jaegerPlatform{
		opt: &option,
	})
	return nil
}

// DEPRECATED: use InitJaeger
func InitOpenTracing(serviceName string, opts ...OptionFunc) error {
	return InitJaeger(serviceName, opts...)
}

type jaegerPlatform struct {
	opt *Option
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
	caller := callerFile + ":" + strconv.Itoa(callerLine)
	span.LogKV("caller", caller)
	return &jaegerTraceImpl{
		ctx:           ctx,
		span:          span,
		operationName: operationName,
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
		ctx:           ctx,
		span:          span,
		operationName: operationName,
		isRoot:        true,
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
		return j.opt.TraceDashboard
	}
	traceID := j.GetTraceID(ctx)
	if traceID == "" {
		return "<disabled>"
	}

	return fmt.Sprintf("%s/%s", j.opt.TraceDashboard, traceID)
}

type jaegerTraceImpl struct {
	ctx           context.Context
	span          opentracing.Span
	operationName string
	isRoot        bool
}

// Context get active context
func (t *jaegerTraceImpl) Context() context.Context {
	return t.ctx
}

// SetTag set tags in tracer span
func (t *jaegerTraceImpl) SetTag(key string, value interface{}) {
	if t.span == nil {
		return
	}

	t.span.SetTag(key, toValue(value))
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

	stackTraces := []string{t.operationName + " => ERROR: " + err.Error()}
	for i := 1; i < 5 && !t.isRoot; i++ {
		_, callerFile, callerLine, _ := runtime.Caller(i)
		if strings.Contains(callerFile, "candi/tracer/jaeger.go") {
			continue
		}
		caller := callerFile + ":" + strconv.Itoa(callerLine)
		stackTraces = append(stackTraces, caller)
	}
	stackTrace := strings.Join(stackTraces, "\n")
	log.Printf("\x1b[31;5m%s\x1b[0m", stackTrace)
	if len(stackTraces) > 1 {
		t.span.LogKV("stacktrace", strings.Join(stackTraces[1:], "\n"))
	}
}

// SetError log data
func (t *jaegerTraceImpl) Log(key string, value interface{}) {
	t.span.LogKV(key, toValue(value))
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

	for k, v := range finishOpt.Tags {
		t.span.SetTag(k, toValue(v))
	}

	if finishOpt.Error != nil {
		t.SetError(finishOpt.Error)
	} else if finishOpt.WithStackTraceDetail {
		stackTraces := []string{t.operationName}
		for i := 1; i < 5; i++ {
			_, callerFile, callerLine, ok := runtime.Caller(i)
			if !ok {
				continue
			}
			if strings.Contains(callerFile, "candi/tracer/jaeger.go") {
				continue
			}
			caller := callerFile + ":" + strconv.Itoa(callerLine)
			stackTraces = append(stackTraces, caller)
		}
		stackTrace := strings.Join(stackTraces, "\n")
		log.Printf("\x1b[32;5m%s\x1b[0m", stackTrace)
		if len(stackTraces) > 1 {
			t.span.LogKV("stacktrace", strings.Join(stackTraces[1:], "\n"))
		}
	}

	t.span.Finish()
}
