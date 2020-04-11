package httpcall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

// TranslatorHTTP implementation
type TranslatorHTTP struct {
	host    string
	httpReq utils.HTTPRequest
}

// NewTranslatorHTTP constructor
func NewTranslatorHTTP() *TranslatorHTTP {
	return &TranslatorHTTP{
		host:    os.Getenv("TRANSLATOR_HOST"),
		httpReq: utils.NewHTTPRequest(10, 50*time.Millisecond),
	}
}

// Translate method
func (b *TranslatorHTTP) Translate(ctx context.Context, from, to, text string) (result string) {

	value := url.Values{}
	value.Set("key", os.Getenv("TRANSLATOR_KEY"))
	value.Set("lang", from+"-"+to)
	value.Add("text", text)

	var url = fmt.Sprintf("%s", b.host)
	body, err := b.httpReq.Do("TranslatorHTTP-Translate", http.MethodPost, url, strings.NewReader(value.Encode()), nil)
	if err != nil {
		return ""
	}

	var response struct {
		Code int      `json:"code"`
		Lang string   `json:"lang"`
		Text []string `json:"text"`
	}

	json.Unmarshal(body, &response)
	return strings.Join(response.Text, " ")
}
