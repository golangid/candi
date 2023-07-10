package ws

import (
	"context"
	"net/http"

	"github.com/golangid/candi/candishared"
	"github.com/gorilla/websocket"
)

const protocolGraphQLWS = "graphql-ws"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{protocolGraphQLWS},
}

// NewHandlerFunc returns an http.HandlerFunc that supports GraphQL over websockets
func NewHandlerFunc(svc GraphQLService, httpHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// handle cors
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE")
		if r.Method == http.MethodOptions {
			return
		}

		for _, subprotocol := range websocket.Subprotocols(r) {
			if subprotocol == protocolGraphQLWS {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				if ws.Subprotocol() != protocolGraphQLWS {
					ws.Close()
					return
				}

				ctx := candishared.SetToContext(context.Background(), candishared.ContextKeyHTTPHeader, r.Header)
				go Connect(ctx, ws, svc)
				return
			}
		}

		// Fallback to HTTP
		httpHandler.ServeHTTP(w, r)
	}
}
