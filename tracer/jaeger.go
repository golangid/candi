package tracer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
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

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(option.agentHost),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		panic(err)
	}

	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.DeploymentEnvironmentKey.String(option.level),
		semconv.TelemetrySDKLanguageGo,
		attribute.Int("num_cpu", runtime.NumCPU()),
		attribute.String("go_version", runtime.Version()),
		attribute.String("candi_version", candi.Version),
	}

	if option.environment != "" {
		attributes = append(attributes, semconv.DeploymentEnvironmentKey.String(option.environment))
	}

	if option.maxGoroutineTag != 0 {
		attributes = append(attributes, attribute.Int("max_goroutines", option.maxGoroutineTag))
	}
	if option.buildNumberTag != "" {
		attributes = append(attributes, attribute.String("build_number", option.buildNumberTag))
	}

	for k, v := range option.attributes {
		attributes = append(attributes, attribute.KeyValue{
			Key: attribute.Key(k), Value: toOtelValue(v),
		})
	}

	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(
			exporter,
			tracesdk.WithMaxExportBatchSize(tracesdk.DefaultMaxExportBatchSize),
			tracesdk.WithBatchTimeout(tracesdk.DefaultScheduleDelay*time.Millisecond),
			tracesdk.WithMaxExportBatchSize(tracesdk.DefaultMaxExportBatchSize),
		),
		tracesdk.WithResource(
			resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		),
	)

	otel.SetTracerProvider(tracerProvider)
	pl := &jaegerPlatform{
		opt:      &option,
		provider: tracerProvider,
		tracer:   tracerProvider.Tracer(serviceName),
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
	opt      *Option
	provider *tracesdk.TracerProvider
	tracer   trace.Tracer
}

func (j *jaegerPlatform) StartSpan(ctx context.Context, operationName string) Tracer {
	ctx, span := j.tracer.Start(ctx, operationName)
	_, callerFile, callerLine, _ := runtime.Caller(4)
	span.AddEvent("", trace.WithAttributes(
		attribute.String("caller", callerFile+":"+strconv.Itoa(callerLine)),
	))
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

	for k, v := range header {
		header[strings.ToLower(k)] = v
	}
	ctx, span := j.tracer.Start(
		propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier(header)),
		operationName,
	)
	span.SetAttributes(attribute.String("trace_id", span.SpanContext().TraceID().String()))
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

	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return span.SpanContext().TraceID().String()
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

func (j *jaegerPlatform) Disconnect(ctx context.Context) error { return j.provider.Shutdown(ctx) }

// jaeger span tracer implementation
type jaegerTraceImpl struct {
	ctx             context.Context
	span            trace.Span
	operationName   string
	isRoot, isPanic bool
	errWhitelist    []error
}

// Context get active context
func (t *jaegerTraceImpl) Context() context.Context {
	return t.ctx
}

// SetTag set tags in tracer span
func (t *jaegerTraceImpl) SetTag(key string, value any) {
	if t.span == nil {
		return
	}

	v, _ := value.(bool)
	t.isPanic = key == "panic" && v
	t.span.SetAttributes(attribute.KeyValue{
		Key: attribute.Key(key), Value: toOtelValue(value),
	})
}

// InjectRequestHeader to continue tracer with custom header carrier
func (t *jaegerTraceImpl) InjectRequestHeader(header map[string]string) {
	if t.span == nil {
		return
	}

	propagation.TraceContext{}.Inject(t.ctx, propagation.MapCarrier(header))
}

// NewContext to continue tracer with new context
func (t *jaegerTraceImpl) NewContext() context.Context {
	return trace.ContextWithSpan(context.Background(), t.span)
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

	t.span.SetStatus(codes.Error, err.Error())

	var stackTraces []string
	for i := 1; i < 10 && len(stackTraces) <= 5 && (!t.isRoot || t.isPanic); i++ {
		if caller := parseCaller(runtime.Caller(i)); caller != "" {
			stackTraces = append(stackTraces, caller)
		}
	}
	t.logStackTrace(31, t.operationName+" => ERROR: "+err.Error(), stackTraces)
}

// Log set log data
func (t *jaegerTraceImpl) Log(key string, value any) {
	t.span.AddEvent("", trace.WithAttributes(
		attribute.KeyValue{
			Key: attribute.Key(key), Value: toOtelValue(value),
		},
	))
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

	for k, v := range finishOpt.Tags {
		t.span.SetAttributes(attribute.KeyValue{
			Key: attribute.Key(k), Value: toOtelValue(v),
		})
	}

	showLogTraceURL := t.isRoot
	if finishOpt.RecoverFunc != nil {
		if rec := recover(); rec != nil {
			finishOpt.RecoverFunc(rec)
			finishOpt.Err = fmt.Errorf("panic: %v", rec)
			t.isRoot = false
			t.span.SetAttributes(attribute.Bool("panic", true))
		}
	}

	if finishOpt.Err != nil {
		t.SetError(finishOpt.Err)
	} else if finishOpt.WithStackTraceDetail {
		var stackTraces []string
		for i := 1; i < 10 && len(stackTraces) <= 5; i++ {
			if caller := parseCaller(runtime.Caller(i)); caller != "" {
				stackTraces = append(stackTraces, caller)
			}
		}
		t.logStackTrace(0, t.operationName, stackTraces)
	}

	if finishOpt.OnFinish != nil {
		finishOpt.OnFinish()
	}
	t.span.End()
	if showLogTraceURL {
		logger.LogGreen(candihelper.ToDelimited(t.operationName, '_') + " > trace_url: " + GetTraceURL(t.ctx))
	}
}

func (t *jaegerTraceImpl) logStackTrace(color int, header string, stackTraces []string) {
	format := "%s"
	if color > 0 {
		format = "\x1b[" + strconv.Itoa(color) + ";5m%s\x1b[0m"
	}
	log.Printf(format, strings.Join(append([]string{header}, stackTraces...), "\n"))
	if len(stackTraces) > 0 {
		t.Log("stacktrace", strings.Join(stackTraces, "\n"))
	}
}
