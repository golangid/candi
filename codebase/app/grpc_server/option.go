package grpcserver

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type (
	option struct {
		tcpPort             string
		debugMode           bool
		jaegerMaxPacketSize int
		sharedListener      cmux.CMux
		serverOptions       []grpc.ServerOption
		tlsConfig           *tls.Config
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		tcpPort:   ":8002",
		debugMode: true,
		serverOptions: []grpc.ServerOption{
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
				PermitWithoutStream: true,            // Allow pings even when there are no active streams
			}),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
				MaxConnectionAgeGrace: 10 * time.Second, // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
				Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
				Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
			}),
			grpc.MaxSendMsgSize(200 * int(candihelper.MByte)), grpc.MaxRecvMsgSize(200 * int(candihelper.MByte)),
		},
	}
}

// SetTCPPort option func
func SetTCPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.tcpPort = fmt.Sprintf(":%d", port)
	}
}

// SetSharedListener option func
func SetSharedListener(sharedListener cmux.CMux) OptionFunc {
	return func(o *option) {
		o.sharedListener = sharedListener
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *option) {
		o.jaegerMaxPacketSize = max
	}
}

// SetServerOptions option func
func SetServerOptions(serverOpts ...grpc.ServerOption) OptionFunc {
	return func(o *option) {
		o.serverOptions = serverOpts
	}
}

// SetTLSConfig option func
func SetTLSConfig(tlsConfig *tls.Config) OptionFunc {
	return func(o *option) {
		o.tlsConfig = tlsConfig
	}
}
