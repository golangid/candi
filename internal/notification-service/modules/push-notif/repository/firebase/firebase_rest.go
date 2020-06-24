package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

type firebaseREST struct {
	host        string
	key         string
	httpRequest utils.HTTPRequest
}

// NewFirebaseREST constructor
func NewFirebaseREST(host, key string) interfaces.PushNotif {
	return &firebaseREST{
		host:        host,
		key:         key,
		httpRequest: utils.NewHTTPRequest(5, 500*time.Millisecond),
	}
}

func (f *firebaseREST) Push(ctx context.Context, req domain.PushRequest) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)

		header := map[string]string{
			"Authorization": fmt.Sprintf("key=%s", f.key),
			"Content-Type":  "application/json",
		}
		body, err := f.httpRequest.Do("FirebaseREST-Push", http.MethodPost, f.host, req, header)
		if err != nil {
			output <- shared.Result{Error: err}
			return
		}

		var resp domain.PushResponse
		json.Unmarshal(body, &resp)
		output <- shared.Result{Data: resp}
	}()

	return output
}
