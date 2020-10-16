package restserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	echoMidd "github.com/labstack/echo/middleware"

	graphqlserver "pkg.agungdwiprasetyo.com/candi/codebase/app/graphql_server"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
	"pkg.agungdwiprasetyo.com/candi/wrapper"
)

type restServer struct {
	serverEngine *echo.Echo
	service      factory.ServiceFactory
	httpPort     string
}

// NewServer create new REST server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {
	server := &restServer{
		serverEngine: echo.New(),
		service:      service,
		httpPort:     fmt.Sprintf(":%d", config.BaseEnv().HTTPPort),
	}

	server.serverEngine.HTTPErrorHandler = wrapper.CustomHTTPErrorHandler
	server.serverEngine.Use(echoMidd.CORS())

	server.serverEngine.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":   fmt.Sprintf("Service %s up and running", service.Name()),
			"timestamp": time.Now().Format(time.RFC3339Nano),
		})
	})

	restRootPath := server.serverEngine.Group("",
		tracer.EchoRestTracerMiddleware, echoMidd.Logger(),
	)
	for _, m := range service.GetModules() {
		if h := m.RestHandler(); h != nil {
			h.Mount(restRootPath)
		}
	}

	httpRoutes := server.serverEngine.Routes()
	sort.Slice(httpRoutes, func(i, j int) bool {
		return httpRoutes[i].Path < httpRoutes[j].Path
	})
	for _, route := range httpRoutes {
		if route.Path != "/" && !strings.Contains(route.Name, "(*Group)") {
			logger.LogGreen(fmt.Sprintf("[REST-ROUTE] %-6s %-30s --> %s", route.Method, route.Path, route.Name))
		}
	}

	// inject graphql handler to rest server
	if config.BaseEnv().UseGraphQL {
		graphqlHandler := graphqlserver.NewHandler(service)
		server.serverEngine.Any("/graphql", echo.WrapHandler(graphqlHandler.ServeGraphQL()))
		server.serverEngine.GET("/graphql/playground", echo.WrapHandler(http.HandlerFunc(graphqlHandler.ServePlayground)))
		server.serverEngine.GET("/graphql/voyager", echo.WrapHandler(http.HandlerFunc(graphqlHandler.ServeVoyager)))

		logger.LogYellow("[GraphQL] endpoint : /graphql")
		logger.LogYellow("[GraphQL] playground : /graphql/playground")
		logger.LogYellow("[GraphQL] voyager : /graphql/voyager")
	}

	fmt.Printf("\x1b[34;1m⇨ HTTP server run at port [::]%s\x1b[0m\n\n", server.httpPort)

	return server
}

func (h *restServer) Serve() {

	h.serverEngine.HideBanner = true
	h.serverEngine.HidePort = true
	if err := h.serverEngine.Start(h.httpPort); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(fmt.Errorf("rest server: %v", e))
		}
	}
}

func (h *restServer) Shutdown(ctx context.Context) {
	log.Println("Stopping REST HTTP server...")
	defer func() { log.Println("Stopping REST HTTP server: \x1b[32;1mSUCCESS\x1b[0m") }()

	if err := h.serverEngine.Shutdown(ctx); err != nil {
		panic(err)
	}
}
