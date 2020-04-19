package delivery

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestProductHandler handler
type RestProductHandler struct {
	mw        middleware.Middleware
	basicAuth echo.MiddlewareFunc
}

// NewRestProductHandler create new rest handler
func NewRestProductHandler(mw middleware.Middleware) *RestProductHandler {
	return &RestProductHandler{
		mw: mw,
		basicAuth: func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set("WWW-Authenticate", `Basic realm=""`)
				if err := mw.BasicAuth(c.Request().Header.Get("Authorization")); err != nil {
					return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
				}

				return next(c)
			}
		},
	}
}

// Mount v1 handler (/v1)
func (h *RestProductHandler) Mount(root *echo.Group) {
	product := root.Group("/product")

	product.GET("", h.getAll, h.basicAuth)
}

func (h *RestProductHandler) getAll(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data product").JSON(c.Response())
}
