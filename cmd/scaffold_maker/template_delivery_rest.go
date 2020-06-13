package main

const deliveryRestTemplate = `package resthandler

import (
	"net/http"

	"{{.PackageName}}/pkg/helper"
	"{{.PackageName}}/pkg/middleware"
	"{{.PackageName}}/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestHandler handler
type RestHandler struct {
	mw middleware.Middleware
}

// NewRestHandler create new rest handler
func NewRestHandler(mw middleware.Middleware) *RestHandler {
	return &RestHandler{
		mw: mw,
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(helper.V1)

	{{clean $.module}} := v1Root.Group("/{{clean $.module}}")
	{{clean $.module}}.GET("", h.hello)
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: {{$.ServiceName}}, module: {{$.module}}").JSON(c.Response())
}

`
