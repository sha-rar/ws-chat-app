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

	// Simple health check.
	mux.HandleFunc("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Test page.
	mux.HandleFunc("/test", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(testPageHTML))
	})

	return mux
}

// inline HTML for quick testing
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
