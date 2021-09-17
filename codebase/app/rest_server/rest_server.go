package restserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo"
	"github.com/soheilhy/cmux"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
)

type restServer struct {
	serverEngine *echo.Echo
	service      factory.ServiceFactory
	listener     net.Listener
	opt          option
}

// NewServer create new REST server
func NewServer(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	server := &restServer{
		serverEngine: echo.New(),
		service:      service,
		opt:          getDefaultOption(),
	}
	for _, opt := range opts {
		opt(&server.opt)
	}

	if server.opt.sharedListener != nil {
		server.listener = server.opt.sharedListener.Match(cmux.HTTP1Fast())
	}

	server.serverEngine.HTTPErrorHandler = CustomHTTPErrorHandler
	server.serverEngine.Use(server.opt.rootMiddlewares...)

	server.serverEngine.GET("/", echo.WrapHandler(server.opt.rootHandler))
	server.serverEngine.GET("/memstats",
		echo.WrapHandler(http.HandlerFunc(candishared.HTTPMemstatsHandler)),
		echo.WrapMiddleware(service.GetDependency().GetMiddleware().HTTPBasicAuth),
	)

	restRootPath := server.serverEngine.Group(server.opt.rootPath, server.echoRestTracerMiddleware)
	if server.opt.debugMode {
		restRootPath.Use(echoLogger())
	}
	for _, m := range service.GetModules() {
		if h := m.RESTHandler(); h != nil {
			h.Mount(restRootPath)
		}
	}

	httpRoutes := server.serverEngine.Routes()
	sort.Slice(httpRoutes, func(i, j int) bool {
		return httpRoutes[i].Path < httpRoutes[j].Path
	})
	for _, route := range httpRoutes {
		if !candihelper.StringInSlice(route.Path, []string{"/", "/memstats"}) && !strings.Contains(route.Name, "(*Group)") {
			logger.LogGreen(fmt.Sprintf("[REST-ROUTE] %-6s %-30s --> %s", route.Method, route.Path, route.Name))
		}
	}

	// inject graphql handler to rest server
	if server.opt.includeGraphQL {
		graphqlHandler := graphqlserver.NewHandler(service, server.opt.graphqlDisableIntrospection)
		server.serverEngine.Any(server.opt.rootPath+"/graphql", echo.WrapHandler(graphqlHandler.ServeGraphQL()))
		server.serverEngine.GET(server.opt.rootPath+"/graphql/playground", echo.WrapHandler(http.HandlerFunc(graphqlHandler.ServePlayground)))
		server.serverEngine.GET(server.opt.rootPath+"/graphql/voyager", echo.WrapHandler(http.HandlerFunc(graphqlHandler.ServeVoyager)))

		logger.LogYellow("[GraphQL] endpoint : " + server.opt.rootPath + "/graphql")
		logger.LogYellow("[GraphQL] playground : " + server.opt.rootPath + "/graphql/playground")
		logger.LogYellow("[GraphQL] voyager : " + server.opt.rootPath + "/graphql/voyager")
	}

	fmt.Printf("\x1b[34;1m⇨ HTTP server run at port [::]%s\x1b[0m\n\n", server.opt.httpPort)

	return server
}

func (h *restServer) Serve() {

	h.serverEngine.HideBanner = true
	h.serverEngine.HidePort = true

	var err error
	if h.listener != nil {
		h.serverEngine.Listener = h.listener
		err = h.serverEngine.Start("")
	} else {
		err = h.serverEngine.Start(h.opt.httpPort)
	}

	switch e := err.(type) {
	case *net.OpError:
		panic(fmt.Errorf("rest server: %v", e))
	}
}

func (h *restServer) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping HTTP server:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	h.serverEngine.Shutdown(ctx)
	if h.listener != nil {
		h.listener.Close()
	}
}

func (h *restServer) Name() string {
	return string(types.REST)
}
