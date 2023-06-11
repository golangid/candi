package restserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golangid/candi/candihelper"
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/wrapper"
	"github.com/soheilhy/cmux"
)

type restServer struct {
	opt        option
	httpEngine *http.Server
	listener   net.Listener
}

// NewServer create new REST server
func NewServer(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	server := &restServer{
		httpEngine: new(http.Server),
		opt:        getDefaultOption(),
	}
	for _, opt := range opts {
		opt(&server.opt)
	}

	mux := chi.NewRouter()
	mux.Use(server.opt.rootMiddlewares...)
	mux.Get("/", server.opt.rootHandler)
	mux.Route("/memstats", func(r chi.Router) {
		r.Use(service.GetDependency().GetMiddleware().HTTPBasicAuth)
		r.Get("/", http.HandlerFunc(wrapper.HTTPHandlerMemstats))
	})
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		wrapper.NewHTTPResponse(http.StatusNotFound, fmt.Sprintf(`Resource "%s %s" not found`, r.Method, r.URL.Path)).JSON(w)
	})

	rootPath := mux.Route(server.opt.rootPath, func(chi.Router) {})
	route := &routeWrapper{router: rootPath}
	for _, m := range service.GetModules() {
		if h := m.RESTHandler(); h != nil {
			h.Mount(route)
		}
	}

	chi.Walk(mux, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if !candihelper.StringInSlice(route, []string{"/", "/memstats/"}) {
			logger.LogGreen(fmt.Sprintf("[REST-ROUTE] %-6s %-30s", method, strings.TrimSuffix(route, "/")))
		}
		return nil
	})

	// inject graphql handler to rest server
	if server.opt.includeGraphQL {
		gqlOpt := server.opt.graphqlOption
		gqlRootPath := gqlOpt.RootPath
		gqlOpt.RootPath = strings.Trim(server.opt.rootPath, "/") + gqlRootPath
		graphqlHandler := graphqlserver.ConstructHandlerFromService(service, gqlOpt)

		rootPath.HandleFunc(gqlRootPath, graphqlHandler.ServeGraphQL())
		rootPath.Get(gqlRootPath+"/playground", http.HandlerFunc(graphqlHandler.ServePlayground))
		rootPath.Get(gqlRootPath+"/voyager", http.HandlerFunc(graphqlHandler.ServeVoyager))
	}

	server.httpEngine.Addr = fmt.Sprintf(":%d", server.opt.httpPort)
	server.httpEngine.Handler = mux

	fmt.Printf("\x1b[34;1m⇨ HTTP server run at port [::]%s\x1b[0m\n\n", server.httpEngine.Addr)

	if server.opt.sharedListener != nil {
		server.listener = server.opt.sharedListener.Match(cmux.HTTP1Fast(http.MethodPatch))
	}

	return server
}

func (s *restServer) Serve() {
	var err error
	if s.listener != nil {
		err = s.httpEngine.Serve(s.listener)
	} else {
		err = s.httpEngine.ListenAndServe()
	}

	switch err.(type) {
	case *net.OpError:
		log.Panicf("REST Server: Unexpected Error: %v", err)
	}
}

func (s *restServer) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping HTTP server:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	s.httpEngine.Shutdown(ctx)
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *restServer) Name() string {
	return string(types.REST)
}
