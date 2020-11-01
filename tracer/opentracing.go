package tracer

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

const maxPacketSize = int(65000 * candihelper.Byte)

// InitOpenTracing with agent and service name
func InitOpenTracing() error {
	serviceName := env.BaseEnv().ServiceName
	if env.BaseEnv().Environment != "" {
		serviceName = fmt.Sprintf("%s-%s", serviceName, strings.ToLower(env.BaseEnv().Environment))
	}

	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  env.BaseEnv().JaegerTracingHost,
		},
		ServiceName: serviceName,
		Tags: []opentracing.Tag{
			{Key: "num_cpu", Value: runtime.NumCPU()},
			{Key: "max_goroutines", Value: env.BaseEnv().MaxGoroutines},
			{Key: "go_version", Value: runtime.Version()},
		},
	}
	tracer, _, err := cfg.NewTracer(config.MaxTagValueLength(math.MaxInt32))
	if err != nil {
		log.Printf("ERROR: cannot init opentracing connection: %v\n", err)
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
