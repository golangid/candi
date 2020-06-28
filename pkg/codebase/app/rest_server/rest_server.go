package restserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
	echoMidd "github.com/labstack/echo/middleware"
)

type restServer struct {
	serverEngine *echo.Echo
	service      factory.ServiceFactory
}

// NewServer create new REST server
func NewServer(service factory.ServiceFactory) factory.AppServerFactory {
	return &restServer{
		serverEngine: echo.New(),
		service:      service,
	}
}

func (h *restServer) Serve() {

	h.serverEngine.HTTPErrorHandler = wrapper.CustomHTTPErrorHandler
	h.serverEngine.Use(echoMidd.Logger(), echoMidd.CORS())

	h.serverEngine.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":   fmt.Sprintf("Service %s up and running", h.service.Name()),
			"timestamp": time.Now().Format(time.RFC3339Nano),
		})
	})

	rootPath := h.serverEngine.Group("")
	for _, m := range h.service.GetModules() {
		if h := m.RestHandler(); h != nil {
			h.Mount(rootPath)
		}
	}

	var routes strings.Builder
	httpRoutes := h.serverEngine.Routes()
	sort.Slice(httpRoutes, func(i, j int) bool {
		return httpRoutes[i].Path < httpRoutes[j].Path
	})
	for _, route := range httpRoutes {
		if !strings.Contains(route.Name, "(*Group)") {
			routeName := strings.TrimPrefix(route.Name, fmt.Sprintf("agungdwiprasetyo.com/backend-microservices/internal/%s/", h.service.Name()))
			routes.WriteString(helper.StringGreen(fmt.Sprintf("[REST-ROUTE] %-6s %-30s --> %s\n", route.Method, route.Path, routeName)))
		}
	}
	fmt.Print(routes.String())

	h.serverEngine.HideBanner = true
	h.serverEngine.HidePort = true
	port := fmt.Sprintf(":%d", config.BaseEnv().RESTPort)
	fmt.Printf("\x1b[34;1m⇨ REST server run at port [::]%s\x1b[0m\n\n", port)
	if err := h.serverEngine.Start(port); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(e)
		}
	}
}

func (h *restServer) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping REST HTTP server...")
	defer deferFunc()

	if err := h.serverEngine.Shutdown(ctx); err != nil {
		panic(err)
	}
}
