package app

import (
	"fmt"
	"log"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// ServeHTTP user service
func (a *App) ServeHTTP() {
	a.httpServer.HTTPErrorHandler = wrapper.CustomHTTPErrorHandler
	a.httpServer.GET("/", func(c echo.Context) error {
		return c.String(200, "Service up and running")
	})

	v1Group := a.httpServer.Group(helper.V1)
	for _, m := range a.modules {
		if h := m.RestHandler(helper.V1); h != nil {
			h.Mount(v1Group)
		}
	}

	if err := a.httpServer.Start(fmt.Sprintf(":%d", config.GlobalEnv.HTTPPort)); err != nil {
		log.Println(err)
	}
}
