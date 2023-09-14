package graphqlserver

import (
	"net/http"
	"strings"

	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/wrapper"
	"github.com/golangid/graphql-go/types"
	"github.com/soheilhy/cmux"
)

type (
	// Option gql server
	Option struct {
		DisableIntrospection bool
		RootPath             string

		httpPort            uint16
		debugMode           bool
		jaegerMaxPacketSize int
		rootHandler         http.Handler
		sharedListener      cmux.CMux
		rootResolver        interfaces.GraphQLHandler
		directiveFuncs      map[string]types.DirectiveFunc
	}

	// OptionFunc type
	OptionFunc func(*Option)
)

func getDefaultOption() Option {
	return Option{
		httpPort:    8000,
		RootPath:    "/graphql",
		debugMode:   true,
		rootHandler: http.HandlerFunc(wrapper.HTTPHandlerDefaultRoot),
	}
}

// SetHTTPPort option func
func SetHTTPPort(port uint16) OptionFunc {
	return func(o *Option) {
		o.httpPort = port
	}
}

// SetRootPath option func
func SetRootPath(rootPath string) OptionFunc {
	return func(o *Option) {
		if strings.Trim(rootPath, "/") == "" {
			return
		}
		o.RootPath = rootPath
	}
}

// SetRootHTTPHandler option func
func SetRootHTTPHandler(rootHandler http.Handler) OptionFunc {
	return func(o *Option) {
		o.rootHandler = rootHandler
	}
}

// SetSharedListener option func
func SetSharedListener(sharedListener cmux.CMux) OptionFunc {
	return func(o *Option) {
		o.sharedListener = sharedListener
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *Option) {
		o.debugMode = debugMode
	}
}

// SetDisableIntrospection option func
func SetDisableIntrospection(disableIntrospection bool) OptionFunc {
	return func(o *Option) {
		o.DisableIntrospection = disableIntrospection
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *Option) {
		o.jaegerMaxPacketSize = max
	}
}

// SetRootResolver option func
func SetRootResolver(resolver interfaces.GraphQLHandler) OptionFunc {
	return func(o *Option) {
		o.rootResolver = resolver
	}
}

// AddDirectiveFunc option func
func AddDirectiveFunc(directiveName string, handlerFunc types.DirectiveFunc) OptionFunc {
	return func(o *Option) {
		if o.directiveFuncs == nil {
			o.directiveFuncs = make(map[string]types.DirectiveFunc)
		}
		o.directiveFuncs[directiveName] = handlerFunc
	}
}
