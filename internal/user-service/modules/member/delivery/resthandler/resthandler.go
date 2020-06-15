package resthandler

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
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

	member := v1Root.Group("/member")
	member.GET("", h.hello)
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: user-service, module: member").JSON(c.Response())
}

