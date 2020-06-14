package delivery

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestProductHandler handler
type RestProductHandler struct {
	mw middleware.Middleware
}

// NewRestProductHandler create new rest handler
func NewRestProductHandler(mw middleware.Middleware) *RestProductHandler {
	return &RestProductHandler{
		mw: mw,
	}
}

// Mount v1 handler (/v1)
func (h *RestProductHandler) Mount(root *echo.Group) {
	product := root.Group("/product")

	product.GET("", h.getAll, h.mw.HTTPBasicAuth(false))
}

func (h *RestProductHandler) getAll(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data product").JSON(c.Response())
}
