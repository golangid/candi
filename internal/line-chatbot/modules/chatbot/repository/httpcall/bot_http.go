package httpcall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

// BotHTTP implementation
type BotHTTP struct {
	host    string
	httpReq utils.HTTPRequest
}

// NewBotHTTP constructor
func NewBotHTTP() *BotHTTP {
	return &BotHTTP{
		host:    os.Getenv("CHATBOT_HOST"),
		httpReq: utils.NewHTTPRequest(10, 50*time.Millisecond),
	}
}

// ProcessText method
func (b *BotHTTP) ProcessText(ctx context.Context, text string) string {
	var url = fmt.Sprintf("%s?input=%s", b.host, url.QueryEscape(text))
	body, err := b.httpReq.Do("Bot-ProcessText", http.MethodGet, url, nil, nil)
	if err != nil {
		return ""
	}

	var output struct {
		Output string `json:"output"`
	}
	json.Unmarshal(body, &output)

	return strings.TrimLeftFunc(output.Output, func(r rune) bool {
		return r == '-' || r == ' '
	})
}

// PushMessageToLine push message to line channel
func (b *BotHTTP) PushMessageToLine(ctx context.Context, message *domain.Message) error {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(true)
	encoder.Encode(message)

	var url = "https://api.line.me/v2/bot/message/push"
	body, err := b.httpReq.Do("Bot-PushMessageToLine", http.MethodPost, url, buffer, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", os.Getenv("LINE_CHANNEL_TOKEN")),
	})
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}
