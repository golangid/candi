package interfaces

import "net/http"

// RESTRouter for REST routing abstraction
type RESTRouter interface {
	Use(middlewares ...func(http.Handler) http.Handler)
	Group(pattern string, middlewares ...func(http.Handler) http.Handler) RESTRouter
	HandleFunc(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	CONNECT(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	DELETE(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	GET(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	HEAD(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	OPTIONS(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	PATCH(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	POST(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	PUT(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
	TRACE(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler)
}
