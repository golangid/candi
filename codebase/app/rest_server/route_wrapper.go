package restserver

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golangid/candi/codebase/interfaces"
)

// URLParam default parse param from url path
var URLParam = func(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

type routeWrapper struct {
	router chi.Router
}

func (r *routeWrapper) Use(middlewares ...func(http.Handler) http.Handler) {
	r.router.Use(middlewares...)
}

func (r *routeWrapper) Group(pattern string, middlewares ...func(http.Handler) http.Handler) interfaces.RESTRouter {
	route := r.router.Route(transformURLParam(pattern), func(chi.Router) {})
	if len(middlewares) > 0 {
		route.Use(middlewares...)
	}
	return &routeWrapper{router: route}
}

func (r *routeWrapper) HandleFunc(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.HandleFunc(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) CONNECT(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Connect(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) DELETE(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Delete(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) GET(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Get(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) HEAD(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Head(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) OPTIONS(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Options(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) PATCH(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Patch(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) POST(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Post(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) PUT(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Put(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func (r *routeWrapper) TRACE(pattern string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.router.Trace(transformURLParam(pattern), WithChainingMiddlewares(h, middlewares...))
}

func transformURLParam(pattern string) string {
	if pattern == "" {
		return "/"
	}
	if strings.ContainsRune(pattern, '{') {
		return pattern
	}

	if strings.ContainsRune(pattern, ':') {
		found := false
		var newPattern strings.Builder
		for i, c := range pattern {
			if c == ':' {
				found = true
				c = '{'
			}
			if found && c == '/' {
				newPattern.WriteRune('}')
				found = false
			}
			newPattern.WriteRune(c)
			if found && i == len(pattern)-1 {
				newPattern.WriteRune('}')
			}
		}
		pattern = newPattern.String()
	}

	return pattern
}
