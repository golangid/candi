package restserver

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/soheilhy/cmux"
)

type (
	option struct {
		rootMiddlewares             []echo.MiddlewareFunc
		rootHandler                 http.Handler
		errorHandler                echo.HTTPErrorHandler
		httpPort                    string
		rootPath                    string
		debugMode                   bool
		includeGraphQL              bool
		graphqlDisableIntrospection bool
		jaegerMaxPacketSize         int
		sharedListener              cmux.CMux
	}

	// OptionFunc type
	OptionFunc func(*option)
)

func getDefaultOption() option {
	return option{
		httpPort:        ":8000",
		rootPath:        "",
		debugMode:       true,
		rootMiddlewares: []echo.MiddlewareFunc{defaultCORS()},
		rootHandler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("REST Server up and running"))
		}),
		errorHandler: CustomHTTPErrorHandler,
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

// SetIncludeGraphQL option func
func SetIncludeGraphQL(includeGraphQL bool) OptionFunc {
	return func(o *option) {
		o.includeGraphQL = includeGraphQL
	}
}

// SetGraphQLDisableIntrospection option func
func SetGraphQLDisableIntrospection(graphqlDisableIntrospection bool) OptionFunc {
	return func(o *option) {
		o.graphqlDisableIntrospection = graphqlDisableIntrospection
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *option) {
		o.jaegerMaxPacketSize = max
	}
}

// SetRootMiddlewares option func
func SetRootMiddlewares(middlewares ...echo.MiddlewareFunc) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = middlewares
	}
}

// AddRootMiddlewares option func
func AddRootMiddlewares(middlewares ...echo.MiddlewareFunc) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = append(o.rootMiddlewares, middlewares...)
	}
}

// SetErrorHandler option func
func SetErrorHandler(errorHandler echo.HTTPErrorHandler) OptionFunc {
	return func(o *option) {
		o.errorHandler = errorHandler
	}
}
