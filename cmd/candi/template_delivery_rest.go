package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.PackageName}}/candihelper"
	"{{.PackageName}}/candishared"
	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/tracer"
	"{{.PackageName}}/wrapper"
)

// RestHandler handler
type RestHandler struct {
	mw        interfaces.Middleware
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewRestHandler create new rest handler
func NewRestHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *RestHandler {
	return &RestHandler{
		mw: mw, uc: uc, validator: validator,
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

	tokenClaim := c.Get(string(candishared.ContextKeyTokenClaim)).(*candishared.TokenClaim) // must using HTTPBearerAuth in middleware for this handler

	return wrapper.NewHTTPResponse(http.StatusOK, h.uc.Hello(ctx) + ", with your session (" + tokenClaim.Audience + ")").JSON(c.Response())
}

`
