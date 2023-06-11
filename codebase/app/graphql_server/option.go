package graphqlserver

import (
	"net/http"

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
		rootResolver        RootResolver
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

// SetRootQuery option func
func SetRootQuery(resolver interface{}) OptionFunc {
	return func(o *Option) {
		o.rootResolver.rootQuery = resolver
	}
}

// SetRootMutation option func
func SetRootMutation(resolver interface{}) OptionFunc {
	return func(o *Option) {
		o.rootResolver.rootMutation = resolver
	}
}

// SetRootSubscription public function
// this public method created because cannot create dynamic method for embedded struct (issue https://github.com/golang/go/issues/15924)
// and subscription in graphql cannot subscribe to at most one subscription at a time
func SetRootSubscription(resolver interface{}) OptionFunc {
	return func(o *Option) {
		o.rootResolver.rootSubscription = resolver
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
