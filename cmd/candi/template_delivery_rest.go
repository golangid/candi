package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

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

	{{clean .ModuleName}} := v1Root.Group("/{{clean .ModuleName}}", echo.WrapMiddleware(h.mw.HTTPBearerAuth))
	{{clean .ModuleName}}.GET("", h.getAll{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.GET("/:id", h.getDetail{{clean (upper .ModuleName)}}ByID, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.POST("", h.save{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.DELETE("/:id", h.delete{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
}

func (h *RestHandler) getAll{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx := trace.Context()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using HTTPBearerAuth in middleware for this handler

	var filter candishared.Filter
	if err := candihelper.ParseFromQueryParam(c.Request().URL.Query(), &filter); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	data, meta, err := h.uc.GetAll{{clean (upper .ModuleName)}}(ctx, filter)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusOK, err.Error()).JSON(c.Response())
	}

	message := "Success, with your user id (" + tokenClaim.Subject + ") and role (" + tokenClaim.Role + ")"
	return wrapper.NewHTTPResponse(http.StatusOK, message, meta, data).JSON(c.Response())
}

func (h *RestHandler) getDetail{{clean (upper .ModuleName)}}ByID(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetDetail{{clean (upper .ModuleName)}}ByID")
	defer trace.Finish()

	data, err := h.uc.GetDetail{{clean (upper .ModuleName)}}(trace.Context(), c.Param("id"))
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success", data).JSON(c.Response())
}

func (h *RestHandler) save{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Save{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	var payload shareddomain.{{clean (upper .ModuleName)}}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusOK, err.Error()).JSON(c.Response())
	}

	err := h.uc.Save{{clean (upper .ModuleName)}}(trace.Context(), &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) delete{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace := tracer.StartTrace(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	if err := h.uc.Delete{{clean (upper .ModuleName)}}(trace.Context(), c.Param("id")); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}
`
