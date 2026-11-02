package http_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/auth"
)

func TestRegisterCreatesUser(t *testing.T) {
	s := newServer(t)

	status, body := s.do(t, http.MethodPost, "/api/auth/register", "", credentials("  alice  ", "secret1"))
	if status != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%s)", status, body)
	}

	var user struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(body), &user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.ID == 0 || user.Username != "alice" {
		t.Errorf("user = %+v, want a new ID and the trimmed username alice", user)
	}

	// The password is stored as a bcrypt hash, never in plain text
	var hash string
	if err := s.DB.QueryRow(`SELECT password_hash FROM users WHERE username = 'alice'`).Scan(&hash); err != nil {
		t.Fatalf("query user: %v", err)
	}
	if hash == "secret1" || bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret1")) != nil {
		t.Errorf("stored password is not a bcrypt hash of the password")
	}
}

func TestRegisterValidatesInput(t *testing.T) {
	s := newServer(t)

	tests := map[string]any{
		"missing username": credentials("", "secret1"),
		"blank username":   credentials("   ", "secret1"),
		"missing password": credentials("alice", ""),
		"short password":   credentials("alice", "12345"),
		"invalid JSON":     "{not json",
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			if status, resBody := s.do(t, http.MethodPost, "/api/auth/register", "", body); status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (%s)", status, resBody)
			}
		})
	}

	if status, _ := s.do(t, http.MethodGet, "/api/auth/register", "", nil); status != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", status)
	}

	var count int
	if err := s.DB.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 0 {
		t.Errorf("%d users created by invalid requests, want 0", count)
	}
}

func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	s := newServer(t)
	s.register(t, "alice", "secret1")

	if status, _ := s.do(t, http.MethodPost, "/api/auth/register", "", credentials("alice", "different1")); status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
}

func TestLoginReturnsWorkingToken(t *testing.T) {
	s := newServer(t)
	s.register(t, "alice", "secret1")

	status, body := s.do(t, http.MethodPost, "/api/auth/login", "", credentials("alice", "secret1"))
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", status, body)
	}
	var res struct {
		Token string `json:"token"`
		User  struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Token == "" || res.User.Username != "alice" {
		t.Fatalf("response = %+v, want a token for alice", res)
	}

	status, body = s.do(t, http.MethodGet, "/api/auth/me", res.Token, nil)
	if status != http.StatusOK {
		t.Fatalf("/me status = %d, want 200 (%s)", status, body)
	}
	var me struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(body), &me); err != nil || me.Username != "alice" {
		t.Errorf("/me = %s, want alice", body)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	s := newServer(t)
	s.register(t, "alice", "secret1")

	wrongPassStatus, wrongPassBody := s.do(t, http.MethodPost, "/api/auth/login", "", credentials("alice", "wrong-password"))
	unknownStatus, unknownBody := s.do(t, http.MethodPost, "/api/auth/login", "", credentials("nobody", "secret1"))

	if wrongPassStatus != http.StatusUnauthorized || unknownStatus != http.StatusUnauthorized {
		t.Errorf("statuses = %d (wrong password), %d (unknown user); want 401 for both", wrongPassStatus, unknownStatus)
	}
	// Identical responses, so login can't be used to discover which usernames exist
	if wrongPassBody != unknownBody {
		t.Errorf("responses differ: %q vs %q", wrongPassBody, unknownBody)
	}
}

func TestAuthenticatedEndpointsRejectBadTokens(t *testing.T) {
	s := newServer(t)
	token := s.registerAndLogin(t, "alice")

	// Same claims, signed with a different secret
	claims, err := auth.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("not-the-server-secret"))
	if err != nil {
		t.Fatalf("sign forged token: %v", err)
	}

	tests := map[string]string{
		"no token":     "",
		"garbage":      "not-a-token",
		"wrong secret": forged,
	}
	for name, tok := range tests {
		t.Run(name, func(t *testing.T) {
			if status, _ := s.do(t, http.MethodGet, "/api/auth/me", tok, nil); status != http.StatusUnauthorized {
				t.Errorf("/me status = %d, want 401", status)
			}
			if tok == "" {
				return // no token is a guest connection, covered elsewhere
			}
			conn, res, err := s.tryDial("general", tok)
			if err == nil {
				conn.Close()
				t.Errorf("websocket accepted a bad token")
			} else if res == nil || res.StatusCode != http.StatusUnauthorized {
				t.Errorf("websocket dial failed with %v, want a 401", err)
			}
		})
	}
}

func TestLogoutRevokesOnlyThatSession(t *testing.T) {
	s := newServer(t)
	s.register(t, "alice", "secret1")
	laptop := s.login(t, "alice", "secret1")
	phone := s.login(t, "alice", "secret1")
	socket := s.dial(t, "general", laptop)

	if status, _ := s.do(t, http.MethodPost, "/api/auth/logout", laptop, nil); status != http.StatusNoContent {
		t.Fatalf("logout status = %d, want 204", status)
	}

	if status, _ := s.do(t, http.MethodGet, "/api/auth/me", laptop, nil); status != http.StatusUnauthorized {
		t.Errorf("logged-out token: /me status = %d, want 401", status)
	}
	if _, res, err := s.tryDial("general", laptop); err == nil || res.StatusCode != http.StatusUnauthorized {
		t.Errorf("logged-out token can still open a websocket")
	}
	if status, _ := s.do(t, http.MethodGet, "/api/auth/me", phone, nil); status != http.StatusOK {
		t.Errorf("other session: /me status = %d, want 200", status)
	}
	if status, _ := s.do(t, http.MethodPost, "/api/auth/logout", laptop, nil); status != http.StatusUnauthorized {
		t.Errorf("second logout status = %d, want 401", status)
	}

	// The socket opened with the revoked session is closed by the server
	_ = socket.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		if _, _, err := socket.conn.ReadMessage(); err != nil {
			if !websocketClosedNormally(err) {
				t.Errorf("socket ended with %v, want a normal close", err)
			}
			break
		}
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	s := newServer(t)
	token := s.registerAndLogin(t, "alice")

	if _, err := s.DB.Exec(`UPDATE sessions SET expires_at = now() - interval '1 minute'`); err != nil {
		t.Fatalf("expire session: %v", err)
	}

	if status, _ := s.do(t, http.MethodGet, "/api/auth/me", token, nil); status != http.StatusUnauthorized {
		t.Errorf("/me status = %d, want 401", status)
	}
}
