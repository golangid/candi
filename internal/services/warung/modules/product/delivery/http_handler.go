package delivery

import (
	"net/http"

	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestProductHandler handler
type RestProductHandler struct {
	midd *middleware.Middleware
}

// NewRestProductHandler create new rest handler
func NewRestProductHandler(midd *middleware.Middleware) *RestProductHandler {
	return &RestProductHandler{
		midd: midd,
	}
}

// Mount v1 handler (/v1)
func (h *RestProductHandler) Mount(root *echo.Group) {
	product := root.Group("/product")

	product.GET("", h.getAll)
}

func (h *RestProductHandler) getAll(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data product").JSON(c.Response())
}
