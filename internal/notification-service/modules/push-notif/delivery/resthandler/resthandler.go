package resthandler

import (
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// RestHandler handler
type RestHandler struct {
	mw interfaces.Middleware
	uc usecase.PushNotifUsecase
}

// NewRestHandler create new rest handler
func NewRestHandler(mw interfaces.Middleware, uc usecase.PushNotifUsecase) *RestHandler {
	return &RestHandler{
		mw: mw,
		uc: uc,
	}
}

// Mount handler with root "/"
// handling version in here
func (h *RestHandler) Mount(root *echo.Group) {
	v1Root := root.Group(helper.V1)

	pushnotif := v1Root.Group("/pushnotif")
	pushnotif.GET("", h.hello)
	pushnotif.POST("/push", h.push, h.mw.HTTPBasicAuth(false))
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: notification-service, module: push-notif").JSON(c.Response())
}

func (h *RestHandler) push(c echo.Context) error {
	if err := h.uc.SendNotification(c.Request().Context()); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed send push notification").JSON(c.Response())
	}
	return wrapper.NewHTTPResponse(http.StatusOK, "Success send push notification").JSON(c.Response())
}
