package http_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gorillaws "github.com/gorilla/websocket"

	"github.com/sha-rar/ws-chat-app/dev/backend/config"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/auth"
	apphttp "github.com/sha-rar/ws-chat-app/dev/backend/internal/http"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/testutil"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

const testSecret = "test-secret-test-secret-test-secret"

type testServer struct {
	*httptest.Server
	DB *sql.DB
}

// newServer starts the full router against a fresh database
func newServer(t *testing.T) *testServer {
	t.Helper()
	conn := testutil.NewDB(t)
	auth.SetSecret(testSecret)

	hub := websocket.NewHub()
	go hub.Run()

	cfg := &config.Config{Env: "development", AllowedOrigins: []string{"http://localhost:3000"}}
	srv := httptest.NewServer(apphttp.NewRouter(cfg, hub))
	t.Cleanup(srv.Close)

	return &testServer{Server: srv, DB: conn}
}

// do sends a request with an optional JSON body and bearer token, returning status and body
func (s *testServer) do(t *testing.T, method, path, token string, body any) (int, string) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, ok := body.(string)
		if !ok {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}
			raw = string(b)
		}
		reader = strings.NewReader(raw)
	}

	req, err := http.NewRequest(method, s.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := s.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return res.StatusCode, buf.String()
}

func credentials(username, password string) map[string]string {
	return map[string]string{"username": username, "password": password}
}

func (s *testServer) register(t *testing.T, username, password string) {
	t.Helper()
	if status, body := s.do(t, http.MethodPost, "/api/auth/register", "", credentials(username, password)); status != http.StatusCreated {
		t.Fatalf("register %s: %d %s", username, status, body)
	}
}

// login returns a token for the user
func (s *testServer) login(t *testing.T, username, password string) string {
	t.Helper()
	status, body := s.do(t, http.MethodPost, "/api/auth/login", "", credentials(username, password))
	if status != http.StatusOK {
		t.Fatalf("login %s: %d %s", username, status, body)
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil || res.Token == "" {
		t.Fatalf("login %s: bad response %q", username, body)
	}
	return res.Token
}

// registerAndLogin creates a user with password "secret1" and returns a token
func (s *testServer) registerAndLogin(t *testing.T, username string) string {
	t.Helper()
	s.register(t, username, "secret1")
	return s.login(t, username, "secret1")
}

// wsEvent is the union of every event the server sends over the socket
type wsEvent struct {
	Type     string          `json:"type"`
	Code     string          `json:"code"`
	Message  json.RawMessage `json:"message"`
	Username string          `json:"username"`
}

type wsClient struct {
	conn    *gorillaws.Conn
	pending []wsEvent
}

// dial opens a socket to room; an empty token connects as a guest
func (s *testServer) dial(t *testing.T, room, token string) *wsClient {
	t.Helper()
	conn, res, err := s.tryDial(room, token)
	if err != nil {
		status := 0
		if res != nil {
			status = res.StatusCode
		}
		t.Fatalf("dial %s: %v (status %d)", room, err, status)
	}
	t.Cleanup(func() { conn.Close() })
	return &wsClient{conn: conn}
}

func (s *testServer) tryDial(room, token string) (*gorillaws.Conn, *http.Response, error) {
	url := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws?roomId=" + room
	if token != "" {
		url += "&token=" + token
	}
	return gorillaws.DefaultDialer.Dial(url, nil)
}

func websocketClosedNormally(err error) bool {
	return gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure)
}

func (c *wsClient) send(t *testing.T, event any) {
	t.Helper()
	if err := c.conn.WriteJSON(event); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// next returns the next event, or ok=false if none arrives within timeout.
// A timed-out read breaks the connection, so only use ok=false as a final check.
func (c *wsClient) next(t *testing.T, timeout time.Duration) (wsEvent, bool) {
	t.Helper()
	if len(c.pending) == 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return wsEvent{}, false
		}
		// The server may batch several events into one frame, one per line
		for _, line := range bytes.Split(data, []byte("\n")) {
			var ev wsEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				t.Fatalf("bad event %q: %v", line, err)
			}
			c.pending = append(c.pending, ev)
		}
	}
	ev := c.pending[0]
	c.pending = c.pending[1:]
	return ev, true
}

// nextOfType skips events until one of the given type arrives
func (c *wsClient) nextOfType(t *testing.T, eventType string) wsEvent {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ev, ok := c.next(t, time.Until(deadline))
		if !ok {
			break
		}
		if ev.Type == eventType {
			return ev
		}
	}
	t.Fatalf("no %q event within 2s", eventType)
	return wsEvent{}
}
