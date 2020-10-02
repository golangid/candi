package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{.PackageName}}/candihelper"
	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/wrapper"
)

// RestHandler handler
type RestHandler struct {
	mw interfaces.Middleware
}

// NewRestHandler create new rest handler
func NewRestHandler(mw interfaces.Middleware) *RestHandler {
	return &RestHandler{
		mw: mw,
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(candihelper.V1)

	{{clean $.module}} := v1Root.Group("/{{clean $.module}}")
	{{clean $.module}}.GET("", h.hello)
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: {{$.ServiceName}}, module: {{$.module}}").JSON(c.Response())
}

`
