package resthandler

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
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
	v1Root := root.Group(helper.V1)

	token := v1Root.Group("/token")
	token.GET("", h.hello)
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: auth-service, module: token").JSON(c.Response())
}

