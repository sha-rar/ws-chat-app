package http

import (
	stdhttp "net/http"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

// NewRouter builds the HTTP router for the app
func NewRouter(hub *websocket.Hub) stdhttp.Handler {
	mux := stdhttp.NewServeMux()

	// WebSocket endpoint: ws://localhost:8080/ws
	mux.HandleFunc("/ws", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Simple health check endpoint
	mux.HandleFunc("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}
