package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.PackageName}}/candihelper"
	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/tracer"
	"{{.PackageName}}/wrapper"
)

// RestHandler handler
type RestHandler struct {
	mw interfaces.Middleware
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewRestHandler create new rest handler
func NewRestHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase) *RestHandler {
	return &RestHandler{
		mw: mw, uc: uc,
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(candihelper.V1)

	{{clean .ModuleName}} := v1Root.Group("/{{clean .ModuleName}}")
	{{clean .ModuleName}}.GET("", h.hello, h.mw.HTTPBearerAuth())
}

func (h *RestHandler) hello(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "DeliveryREST:Hello")
	defer trace.Finish()
	ctx := trace.Context()

	return wrapper.NewHTTPResponse(http.StatusOK, h.uc.Hello(ctx)).JSON(c.Response())
}

`
