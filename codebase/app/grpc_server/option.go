package grpcserver

import (
	"fmt"

	"github.com/soheilhy/cmux"
)

type (
	option struct {
		tcpPort             string
		debugMode           bool
		jaegerMaxPacketSize int
		sharedListener      cmux.CMux
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		tcpPort:   ":8002",
		debugMode: true,
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
