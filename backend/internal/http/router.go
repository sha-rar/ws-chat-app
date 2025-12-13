// backend/internal/http/router.go
package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

func NewRouter(hub *websocket.Hub) http.Handler {
	mux := http.NewServeMux()

	// WebSocket
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Auth
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)

	// Message history
	mux.HandleFunc("/api/rooms/", handleRoomMessages)

	// Health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Test page
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(testPageHTML))
	})

	return withCORS(mux)
}

// GET /api/rooms/{roomId}/messages?limit=50
func handleRoomMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "messages" {
		http.NotFound(w, r)
		return
	}
	roomID := parts[0]
	if roomID == "" {
		http.Error(w, "roomId is required", http.StatusBadRequest)
		return
	}

	limit := 50
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if parsed, err := strconv.Atoi(ls); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	msgs := store.GetRecentMessages(roomID, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		http.Error(w, "failed to encode messages", http.StatusInternalServerError)
		return
	}
}

// Basic CORS wrapper for dev
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Inline HTML for quick testing
const testPageHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>WS Test</title>
    <style>
      body { font-family: sans-serif; padding: 16px; }
      #log { border: 1px solid #ccc; height: 200px; overflow-y: auto; padding: 8px; margin-bottom: 8px; }
    </style>
  </head>
  <body>
    <h1>WebSocket Test</h1>
    <div id="log"></div>
    <input id="msg" placeholder="Type a message" />
    <button id="send">Send</button>

    <script>
      const logEl = document.getElementById("log");
      const inputEl = document.getElementById("msg");
      const sendBtn = document.getElementById("send");

      const params = new URLSearchParams(window.location.search);
      const roomId = params.get("roomId") || "general";
      const username = "test-user"; // hardcoded for now

      function log(msg) {
        const div = document.createElement("div");
        div.textContent = msg;
        logEl.appendChild(div);
        logEl.scrollTop = logEl.scrollHeight;
      }

      const ws = new WebSocket(
        "ws://localhost:8080/ws?roomId=" + encodeURIComponent(roomId)
      );

      ws.onopen = () => log('Connected to room "' + roomId + '"');
      ws.onclose = () => log("Disconnected");
      ws.onerror = (e) => log("Error: " + e.message);

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          log(
            "[" +
              msg.roomId +
              "] " +
              msg.username +
              ": " +
              msg.text +
              " (" +
              msg.timestamp +
              ")"
          );
        } catch (e) {
          log("(raw) " + event.data);
        }
      };

      sendBtn.onclick = () => {
        const text = inputEl.value;
        if (!text) return;

        const payload = {
          username,
          roomId,
          text,
        };

        ws.send(JSON.stringify(payload));
        log("You: " + text);
        inputEl.value = "";
      };

      inputEl.addEventListener("keydown", (e) => {
        if (e.key === "Enter") sendBtn.click();
      });
    </script>
  </body>
</html>
`
