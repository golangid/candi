package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
	echoMidd "github.com/labstack/echo/middleware"
)

// ServeHTTP user service
func (a *App) ServeHTTP() {
	a.httpServer.HTTPErrorHandler = wrapper.CustomHTTPErrorHandler
	a.httpServer.Use(echoMidd.Logger())

	a.httpServer.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":   "Service up and running",
			"timestamp": time.Now().Format(time.RFC3339Nano),
		})
	})

	v1Group := a.httpServer.Group(helper.V1)
	for _, m := range a.modules {
		if h := m.RestHandler(helper.V1); h != nil {
			h.Mount(v1Group)
		}
	}

	var routes strings.Builder
	for _, route := range a.httpServer.Routes() {
		if !strings.Contains(route.Name, "(*Group)") {
			routes.WriteString(helper.StringGreen(fmt.Sprintf("[ROUTE] %-7s %-30s --> %s\n", route.Method, route.Path, route.Name)))
		}
	}
	fmt.Print(routes.String())

	if err := a.httpServer.Start(fmt.Sprintf(":%d", config.GlobalEnv.HTTPPort)); err != nil {
		log.Println(err)
	}
}
