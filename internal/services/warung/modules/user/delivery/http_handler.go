package delivery

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestUserHandler handler
type RestUserHandler struct {
	mw middleware.Middleware
}

// NewRestUserHandler create new rest handler
func NewRestUserHandler(mw middleware.Middleware) *RestUserHandler {
	return &RestUserHandler{
		mw: mw,
	}
}

// Mount v1 handler (/v1)
func (h *RestUserHandler) Mount(root *echo.Group) {
	user := root.Group("/user")

	user.GET("", h.getAll)
}

func (h *RestUserHandler) getAll(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data user").JSON(c.Response())
}
