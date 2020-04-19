package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
)

// RestHandler handler
type RestHandler struct {
	mw        middleware.Middleware
	uc        usecase.BotUsecase
	basicAuth echo.MiddlewareFunc
}

// NewRestHandler create new rest handler
func NewRestHandler(mw middleware.Middleware, uc usecase.BotUsecase) *RestHandler {
	return &RestHandler{
		mw: mw,
		uc: uc,
		basicAuth: func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if err := mw.BasicAuth(c.Request().Header.Get("Authorization")); err != nil {
					return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
				}

				return next(c)
			}
		},
	}
}

// Mount v1 handler (/v1)
func (h *RestHandler) Mount(root *echo.Group) {
	bot := root.Group("/bot")

	bot.POST("/callback", h.callback)
	bot.POST("/pushmessage", h.pushMessage, h.basicAuth)
}

func (h *RestHandler) callback(c echo.Context) error {

	req := c.Request()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	signature := req.Header.Get("X-Line-Signature")
	fmt.Println("signature :==>", signature)
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
	}

	hash := hmac.New(sha256.New, []byte(os.Getenv("LINE_CHANNEL_SECRET")))
	_, err = hash.Write(body)
	if err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	if !hmac.Equal(decoded, hash.Sum(nil)) {
		return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
	}

	request := struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err = json.Unmarshal(body, &request); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

	if err := h.uc.ProcessCallback(c.Request().Context(), request.Events); err != nil {
		return wrapper.NewHTTPResponse(http.StatusBadRequest, err.Error()).JSON(c.Response())
	}

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
