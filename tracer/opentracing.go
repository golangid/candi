package tracer

import (
	"log"
	"math"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
	"pkg.agungdwiprasetyo.com/gendon/helper"
)

const maxPacketSize = int(65000 * helper.Byte)

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
