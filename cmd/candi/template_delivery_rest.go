package main

const deliveryRestTemplate = `// {{.Header}}

package resthandler

import (
	"net/http"

	"github.com/labstack/echo"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
	"{{.LibraryName}}/wrapper"
)

// RestHandler handler
type RestHandler struct {
	mw        interfaces.Middleware
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRestHandler create new rest handler
func NewRestHandler(uc usecase.Usecase, deps dependency.Dependency) *RestHandler {
	return &RestHandler{
		uc: uc, mw: deps.GetMiddleware(), validator: deps.GetValidator(),
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(candihelper.V1)

	{{clean .ModuleName}} := v1Root.Group("/{{clean .ModuleName}}", echo.WrapMiddleware(h.mw.HTTPBearerAuth))
	{{clean .ModuleName}}.GET("", h.getAll{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.GET("/:id", h.getDetail{{clean (upper .ModuleName)}}ByID, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.POST("", h.create{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.PUT("/:id", h.update{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
	{{clean .ModuleName}}.DELETE("/:id", h.delete{{clean (upper .ModuleName)}}, echo.WrapMiddleware(h.mw.HTTPPermissionACL("resource.public")))
}

func (h *RestHandler) getAll{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using HTTPBearerAuth in middleware for this handler

	var filter domain.Filter{{clean (upper .ModuleName)}}
	if err := candihelper.ParseFromQueryParam(c.Request().URL.Query(), &filter); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	data, meta, err := h.uc.{{clean (upper .ModuleName)}}().GetAll{{clean (upper .ModuleName)}}(ctx, &filter)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusOK, err.Error()).JSON(c.Response())
	}

	message := "Success, with your user id (" + tokenClaim.Subject + ") and role (" + tokenClaim.Role + ")"
	return wrapper.NewHTTPResponse(http.StatusOK, message, meta, data).JSON(c.Response())
}

func (h *RestHandler) getDetail{{clean (upper .ModuleName)}}ByID(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:GetDetail{{clean (upper .ModuleName)}}ByID")
	defer trace.Finish()

	data, err := h.uc.{{clean (upper .ModuleName)}}().GetDetail{{clean (upper .ModuleName)}}(ctx, c.Param("id"))
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success", data).JSON(c.Response())
}

func (h *RestHandler) create{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	var payload shareddomain.{{clean (upper .ModuleName)}}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusOK, err.Error()).JSON(c.Response())
	}

	err := h.uc.{{clean (upper .ModuleName)}}().Create{{clean (upper .ModuleName)}}(ctx, &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) update{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	var payload shareddomain.{{clean (upper .ModuleName)}}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusOK, err.Error()).JSON(c.Response())
	}

	err := h.uc.{{clean (upper .ModuleName)}}().Update{{clean (upper .ModuleName)}}(ctx, c.Param("id"), &payload)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}

func (h *RestHandler) delete{{clean (upper .ModuleName)}}(c echo.Context) error {
	trace, ctx := tracer.StartTraceWithContext(c.Request().Context(), "{{clean (upper .ModuleName)}}DeliveryREST:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	if err := h.uc.{{clean (upper .ModuleName)}}().Delete{{clean (upper .ModuleName)}}(ctx, c.Param("id")); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "Success").JSON(c.Response())
}
`
