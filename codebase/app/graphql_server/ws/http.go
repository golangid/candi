package ws

import (
	"context"
	"net/http"
	"slices"

	"github.com/golangid/candi/candishared"
	"github.com/gorilla/websocket"
)

// ConnectOption gql ws server
type ConnectOption struct {
	bufferSize int
	readLimit  int64
}

// SetBufferSize func
func SetBufferSize(size int) func(*ConnectOption) {
	return func(conn *ConnectOption) {
		conn.bufferSize = size
	}
}

// SetReadLimit limits the maximum size of incoming messages
func SetReadLimit(limit int64) func(*ConnectOption) {
	return func(conn *ConnectOption) {
		conn.readLimit = limit
	}
}

// NewHandlerFunc returns an http.HandlerFunc that supports GraphQL over websockets
func NewHandlerFunc(svc GraphQLService, httpHandler http.Handler, connectOptions ...func(*ConnectOption)) http.HandlerFunc {
	connectOption := ConnectOption{
		bufferSize: 1024,
		readLimit:  1<<16 - 1, // 65535
	}
	for _, opt := range connectOptions {
		opt(&connectOption)
	}
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  connectOption.bufferSize,
		WriteBufferSize: connectOption.bufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: []string{"graphql-ws", "graphql-transport-ws"},
	}

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
			if slices.Contains(upgrader.Subprotocols, subprotocol) {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}
				ws.SetReadLimit(connectOption.readLimit)

				if !slices.Contains(upgrader.Subprotocols, ws.Subprotocol()) {
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
