package delivery

import (
	"net/http"
	"sync"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

var someMap = sync.Map{}

// RestInvitationHandler handler
type RestInvitationHandler struct {
	mw middleware.Middleware
}

// NewRestInvitationHandler create new rest handler
func NewRestInvitationHandler(mw middleware.Middleware) *RestInvitationHandler {
	return &RestInvitationHandler{
		mw: mw,
	}
}

// Mount v1 handler (/v1)
func (h *RestInvitationHandler) Mount(root *echo.Group) {
	invitation := root.Group("/invitation")

	invitation.GET("", h.getAll)
}

func (h *RestInvitationHandler) getAll(c echo.Context) error {
	id, add := c.QueryParam("id"), c.QueryParam("add")
	// debug.Println(id, add)

	if _, ok := someMap.Load(id); ok {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Masih kelock yg "+add).JSON(c.Response())
	}

	someMap.Store(id, true)
	defer func() {
		someMap.Delete(id)
	}()

	// debug.Println("proses yg ", add)
	time.Sleep(5 * time.Second)

	return wrapper.NewHTTPResponse(http.StatusOK, "Sukses mengambil data invitation "+add).JSON(c.Response())
}
