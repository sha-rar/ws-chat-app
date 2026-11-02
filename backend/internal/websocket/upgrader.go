package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// NewUpgrader returns an upgrader that only accepts connections from the
// given origins, so other sites can't open sockets on a visitor's behalf
func NewUpgrader(allowedOrigins []string) *websocket.Upgrader {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Non-browser clients (no Origin header) aren't subject to CSRF
			if origin == "" {
				return true
			}
			return allowed[strings.TrimRight(origin, "/")]
		},
	}
}
