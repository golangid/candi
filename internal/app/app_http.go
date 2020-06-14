package app

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
	echoMidd "github.com/labstack/echo/middleware"
)

// ServeHTTP service
func (a *App) ServeHTTP() {
	if a.httpServer == nil {
		return
	}

	a.httpServer.HTTPErrorHandler = wrapper.CustomHTTPErrorHandler
	a.httpServer.Use(echoMidd.Logger(), echoMidd.CORS())

	a.httpServer.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":   fmt.Sprintf("Service %s up and running", a.service.Name()),
			"timestamp": time.Now().Format(time.RFC3339Nano),
		})
	})

	rootPath := a.httpServer.Group("")
	for _, m := range a.service.GetModules() {
		if h := m.RestHandler(); h != nil {
			h.Mount(rootPath)
		}
	}

	var routes strings.Builder
	httpRoutes := a.httpServer.Routes()
	sort.Slice(httpRoutes, func(i, j int) bool {
		return httpRoutes[i].Path < httpRoutes[j].Path
	})
	for _, route := range httpRoutes {
		if !strings.Contains(route.Name, "(*Group)") {
			routes.WriteString(helper.StringGreen(fmt.Sprintf("[HTTP-ROUTE] %-6s %-30s --> %s\n", route.Method, route.Path, route.Name)))
		}
	}
	fmt.Print(routes.String())

	a.httpServer.HideBanner = true
	if err := a.httpServer.Start(fmt.Sprintf(":%d", config.BaseEnv().HTTPPort)); err != nil {
		log.Println(err)
	}
}
