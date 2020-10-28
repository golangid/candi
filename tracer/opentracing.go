package tracer

import (
	"log"
	"math"
	"runtime"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	candiconfig "pkg.agungdwiprasetyo.com/candi/config"
)

const maxPacketSize = int(65000 * candihelper.Byte)

var agent string

// InitOpenTracing with agent and service name
func InitOpenTracing(agentHost, serviceName string) error {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHost,
		},
		ServiceName: serviceName,
		Tags: []opentracing.Tag{
			{Key: "num_cpu", Value: runtime.NumCPU()},
			{Key: "max_goroutines", Value: candiconfig.BaseEnv().MaxGoroutines},
			{Key: "go_version", Value: runtime.Version()},
		},
	}
	tracer, _, err := cfg.NewTracer(config.MaxTagValueLength(math.MaxInt32))
	if err != nil {
		log.Printf("ERROR: cannot init opentracing connection: %v\n", err)
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	agent = agentHost
	return nil
}
