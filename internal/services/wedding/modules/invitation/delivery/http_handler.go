package delivery

import (
	"net/http"

	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestInvitationHandler handler
type RestInvitationHandler struct {
	midd *middleware.Middleware
}

// NewRestInvitationHandler create new rest handler
func NewRestInvitationHandler(midd *middleware.Middleware) *RestInvitationHandler {
	return &RestInvitationHandler{
		midd: midd,
	}
}

// Mount v1 handler (/v1)
func (h *RestInvitationHandler) Mount(root *echo.Group) {
	invitation := root.Group("/invitation")

	invitation.GET("", h.getAll)
}

func (h *RestInvitationHandler) getAll(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data invitation").JSON(c.Response())
}
