package resthandler

import (
	"net/http"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
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
	pushnotif.POST("/schedule", h.scheduledNotification, h.mw.HTTPBasicAuth(false))
}

func (h *RestHandler) hello(c echo.Context) error {
	return wrapper.NewHTTPResponse(http.StatusOK, "Hello, from service: notification-service, module: push-notif").JSON(c.Response())
}

func (h *RestHandler) push(c echo.Context) error {
	var payload domain.PushNotifRequestPayload
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed parse body payload", err).JSON(c.Response())
	}

	if err := h.uc.SendNotification(c.Request().Context(), &payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed send push notification").JSON(c.Response())
	}
	return wrapper.NewHTTPResponse(http.StatusOK, "Success send push notification").JSON(c.Response())
}

func (h *RestHandler) scheduledNotification(c echo.Context) error {
	var payload struct {
		ScheduledAt string                         `json:"scheduledAt"`
		Data        domain.PushNotifRequestPayload `json:"data"`
	}
	if err := c.Bind(&payload); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed parse body payload", err).JSON(c.Response())
	}

	scheduledAt, err := time.Parse(time.RFC3339, payload.ScheduledAt)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed parse scheduled time format", err).JSON(c.Response())
	}

	if scheduledAt.Before(time.Now()) {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Scheduled time must in future").JSON(c.Response())
	}

	if err := h.uc.SendScheduledNotification(c.Request().Context(), scheduledAt, &payload.Data); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, "Failed set scheduled push notification").JSON(c.Response())
	}
	return wrapper.NewHTTPResponse(http.StatusOK, "Success set scheduled push notification, scheduled at: "+payload.ScheduledAt).JSON(c.Response())
}
