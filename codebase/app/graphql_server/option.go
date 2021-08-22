package graphqlserver

import (
	"fmt"
	"net/http"

	"github.com/soheilhy/cmux"
)

type (
	option struct {
		httpPort             string
		rootPath             string
		debugMode            bool
		disableIntrospection bool
		jaegerMaxPacketSize  int
		rootHandler          http.Handler
		sharedListener       cmux.CMux
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		httpPort:  ":8000",
		rootPath:  "",
		debugMode: true,
		rootHandler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("REST Server up and running"))
		}),
	}
}

// SetHTTPPort option func
func SetHTTPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.httpPort = fmt.Sprintf(":%d", port)
	}
}

// SetRootPath option func
func SetRootPath(rootPath string) OptionFunc {
	return func(o *option) {
		o.rootPath = rootPath
	}
}

// SetRootHTTPHandler option func
func SetRootHTTPHandler(rootHandler http.Handler) OptionFunc {
	return func(o *option) {
		o.rootHandler = rootHandler
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

// SetDisableIntrospection option func
func SetDisableIntrospection(disableIntrospection bool) OptionFunc {
	return func(o *option) {
		o.disableIntrospection = disableIntrospection
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *option) {
		o.jaegerMaxPacketSize = max
	}
}
