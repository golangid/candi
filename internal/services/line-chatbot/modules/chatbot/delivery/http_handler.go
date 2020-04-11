package delivery

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
)

// RestHandler handler
type RestHandler struct {
	mw         middleware.Middleware
	uc         usecase.BotUsecase
	lineClient *linebot.Client
}

// NewRestHandler create new rest handler
func NewRestHandler(mw middleware.Middleware, client *linebot.Client, uc usecase.BotUsecase) *RestHandler {
	return &RestHandler{
		mw:         mw,
		lineClient: client,
		uc:         uc,
	}
}

// Mount v1 handler (/v1)
func (h *RestHandler) Mount(root *echo.Group) {
	bot := root.Group("/bot")

	bot.POST("/callback", h.callback)
	bot.POST("/pushmessage", h.pushMessage)
}

func (h *RestHandler) callback(c echo.Context) error {

	req := c.Request()
	events, err := h.lineClient.ParseRequest(req)
	if err != nil {
		var code int
		if err == linebot.ErrInvalidSignature {
			code = http.StatusBadRequest
		} else {
			code = http.StatusInternalServerError
		}

		return wrapper.NewHTTPResponse(code, err.Error()).JSON(c.Response())
	}

	h.uc.ProcessCallback(c.Request().Context(), events)
	return wrapper.NewHTTPResponse(http.StatusOK, "ok").JSON(c.Response())
}

func (h *RestHandler) pushMessage(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	var message struct {
		To      string `json:"to"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err = json.Unmarshal(body, &message); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	if err := h.uc.PushMessageToChannel(c.Request().Context(), message.To, message.Title, message.Message); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	return wrapper.NewHTTPResponse(http.StatusOK, "ok").JSON(c.Response())
}
