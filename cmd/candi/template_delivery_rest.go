package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
	"{{.LibraryName}}/wrapper"
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
	{{clean .ModuleName}}.GET("", h.hello, echo.WrapMiddleware(h.mw.HTTPBearerAuth))
}

func (h *RestHandler) hello(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Hello")
	defer trace.Finish()
	ctx := trace.Context()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using HTTPBearerAuth in middleware for this handler

	return wrapper.NewHTTPResponse(http.StatusOK, h.uc.Hello(ctx) + ", with your session (" + tokenClaim.Audience + ")").JSON(c.Response())
}
`
