package tracer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/gomodule/redigo/redis"
	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InitJaeger init jaeger tracing
func InitJaeger(serviceName string, opts ...OptionFunc) PlatformType {
	option := Option{
		agentHost:       env.BaseEnv().JaegerTracingHost,
		level:           env.BaseEnv().Environment,
		buildNumberTag:  env.BaseEnv().BuildNumber,
		maxGoroutineTag: env.BaseEnv().MaxGoroutines,
		errorWhitelist:  []error{redis.ErrNil, sql.ErrNoRows, mongo.ErrNoDocuments},
	}
	urlAgent, err := url.Parse("//" + env.BaseEnv().JaegerTracingHost)
	if urlAgent != nil && err == nil {
		option.traceDashboard = fmt.Sprintf("http://%s:16686/trace", urlAgent.Hostname())
	}

	for _, opt := range opts {
		opt(&option)
	}

	if option.level != "" {
		serviceName = fmt.Sprintf("%s-%s", serviceName, strings.ToLower(option.level))
	}
	defaultTags := []opentracing.Tag{
		{Key: "num_cpu", Value: runtime.NumCPU()},
		{Key: "go_version", Value: runtime.Version()},
		{Key: "candi_version", Value: candi.Version},
	}
	if option.maxGoroutineTag != 0 {
		defaultTags = append(defaultTags, opentracing.Tag{
			Key: "max_goroutines", Value: option.maxGoroutineTag,
		})
	}
	if option.buildNumberTag != "" {
		defaultTags = append(defaultTags, opentracing.Tag{
			Key: "build_number", Value: option.buildNumberTag,
		})
	}
	cfg := &jaeger.Configuration{
		Sampler: &jaeger.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaeger.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  option.agentHost,
		},
		ServiceName: serviceName,
		Tags:        defaultTags,
	}
	tracer, closer, err := cfg.NewTracer(jaeger.MaxTagValueLength(math.MaxInt32))
	if err != nil {
		log.Panicf("ERROR: cannot init jaeger opentracing connection: %v\n", err)
	}
	opentracing.SetGlobalTracer(tracer)

	pl := &jaegerPlatform{
		opt:    &option,
		closer: closer,
	}
	SetTracerPlatformType(pl)
	return pl
}

// DEPRECATED: use InitJaeger
func InitOpenTracing(serviceName string, opts ...OptionFunc) error {
	InitJaeger(serviceName, opts...)
	return nil
}

// jaeger platform
type jaegerPlatform struct {
	opt    *Option
	closer io.Closer
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
	if j.opt.logAllSpan {
		_, callerFile, callerLine, _ := runtime.Caller(3)
		log.Printf("\x1b[32;5m%s => %s:%d\x1b[0m", operationName, callerFile, callerLine)
	}
	return &jaegerTraceImpl{
		ctx:           ctx,
		span:          span,
		operationName: operationName,
		errWhitelist:  j.opt.errorWhitelist,
	}
}

func (j *jaegerPlatform) StartRootSpan(ctx context.Context, operationName string, header map[string]string) Tracer {
	if header == nil {
		header = make(map[string]string)
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
	span.SetTag("trace_id", j.GetTraceID(ctx))
	return &jaegerTraceImpl{
		ctx:           ctx,
		span:          span,
		operationName: operationName,
		isRoot:        true,
		errWhitelist:  j.opt.errorWhitelist,
	}
}

func (j *jaegerPlatform) GetTraceID(ctx context.Context) string {
	if j.opt.traceIDExtractor != nil {
		return j.opt.traceIDExtractor(ctx)
	}

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}

	stringer, ok := span.(fmt.Stringer)
	if !ok {
		return ""
	}
	traceID := stringer.String()
	splits := strings.Split(traceID, ":")
	if len(splits) > 0 {
		return splits[0]
	}

	return traceID
}

func (j *jaegerPlatform) GetTraceURL(ctx context.Context) (u string) {
	if ctx == nil {
		return j.opt.traceDashboard
	}
	traceID := j.GetTraceID(ctx)
	if traceID == "" {
		return "<disabled>"
	}

	return fmt.Sprintf("%s/%s", j.opt.traceDashboard, traceID)
}

func (j *jaegerPlatform) Disconnect(ctx context.Context) error { return j.closer.Close() }

// jaeger span tracer implementation
type jaegerTraceImpl struct {
	ctx           context.Context
	span          opentracing.Span
	operationName string
	isRoot        bool
	errWhitelist  []error
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
	for _, errWhitelist := range t.errWhitelist {
		if errors.Is(errWhitelist, err) {
			t.Log("error", err.Error())
			return
		}
	}

	ext.Error.Set(t.span, true)
	t.span.LogKV("error.message", err.Error())

	var stackTraces []string
	for i := 1; i < 10 && len(stackTraces) <= 5 && !t.isRoot; i++ {
		if caller := parseCaller(runtime.Caller(i)); caller != "" {
			stackTraces = append(stackTraces, caller)
		}
	}
	t.logStackTrace(31, t.operationName+" => ERROR: "+err.Error(), stackTraces)
}

// Log set log data
func (t *jaegerTraceImpl) Log(key string, value interface{}) {
	t.span.LogKV(key, toValue(value))
}

// Finish trace must in deferred function
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

	for k, v := range finishOpt.tags {
		t.span.SetTag(k, toValue(v))
	}

	showLogTraceURL := t.isRoot
	if finishOpt.recoverFunc != nil {
		if rec := recover(); rec != nil {
			finishOpt.recoverFunc(rec)
			finishOpt.err = fmt.Errorf("panic: %v", rec)
			t.isRoot = false
			t.span.SetTag("panic", true)
		}
	}

	if finishOpt.err != nil {
		t.SetError(finishOpt.err)
	} else if finishOpt.withStackTraceDetail {
		var stackTraces []string
		for i := 1; i < 10 && len(stackTraces) <= 5; i++ {
			if caller := parseCaller(runtime.Caller(i)); caller != "" {
				stackTraces = append(stackTraces, caller)
			}
		}
		t.logStackTrace(32, t.operationName, stackTraces)
	}

	if finishOpt.onFinish != nil {
		finishOpt.onFinish()
	}
	t.span.Finish()
	if showLogTraceURL {
		logger.LogGreen(candihelper.ToDelimited(t.operationName, '_') + " > trace_url: " + GetTraceURL(t.ctx))
	}
}

func (t *jaegerTraceImpl) logStackTrace(color int, header string, stackTraces []string) {
	log.Printf("\x1b[%d;5m%s\x1b[0m", color, strings.Join(append([]string{header}, stackTraces...), "\n"))
	if len(stackTraces) > 0 {
		t.span.LogKV("stacktrace", strings.Join(stackTraces, "\n"))
	}
}
