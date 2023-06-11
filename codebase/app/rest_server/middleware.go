package restserver

import "net/http"

// WithChainingMiddlewares chaining middlewares
func WithChainingMiddlewares(handlerFunc http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.HandlerFunc {
	if len(middlewares) == 0 {
		return handlerFunc
	}

	wrapped := handlerFunc
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](http.HandlerFunc(wrapped)).ServeHTTP
	}
	return wrapped
}
