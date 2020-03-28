package delivery

import (
	"net/http"

	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestUserHandler handler
type RestUserHandler struct {
	midd *middleware.Middleware
}

// NewRestUserHandler create new rest handler
func NewRestUserHandler(midd *middleware.Middleware) *RestUserHandler {
	return &RestUserHandler{
		midd: midd,
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
