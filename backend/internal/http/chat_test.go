package http_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
)

// settle gives the hub a moment to register sockets that were just dialed
func settle() {
	time.Sleep(100 * time.Millisecond)
}

func decodeChat(t *testing.T, ev wsEvent) models.Message {
	t.Helper()
	var msg models.Message
	if err := json.Unmarshal(ev.Message, &msg); err != nil {
		t.Fatalf("decode chat message: %v", err)
	}
	return msg
}

func (s *testServer) history(t *testing.T, room, query string) []models.Message {
	t.Helper()
	status, body := s.do(t, http.MethodGet, "/api/rooms/"+room+"/messages"+query, "", nil)
	if status != http.StatusOK {
		t.Fatalf("history status = %d, want 200 (%s)", status, body)
	}
	var msgs []models.Message
	if err := json.Unmarshal([]byte(body), &msgs); err != nil {
		t.Fatalf("decode history %q: %v", body, err)
	}
	return msgs
}

func (s *testServer) messageCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.DB.QueryRow(`SELECT count(*) FROM messages`).Scan(&n); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	return n
}

func TestChatMessageIsBroadcastPersistedAndInHistory(t *testing.T) {
	s := newServer(t)
	alice := s.dial(t, "general", s.registerAndLogin(t, "alice"))
	bob := s.dial(t, "general", s.registerAndLogin(t, "bob"))
	elsewhere := s.dial(t, "random", s.registerAndLogin(t, "carol"))
	settle()

	// The username field is ignored; the server uses the authenticated user
	alice.send(t, map[string]string{"type": "chat", "text": "  hello bob  ", "username": "mallory"})

	for name, c := range map[string]*wsClient{"alice": alice, "bob": bob} {
		msg := decodeChat(t, c.nextOfType(t, "chat"))
		if msg.ID == 0 || msg.Username != "alice" || msg.Text != "hello bob" || msg.RoomID != "general" {
			t.Errorf("%s received %+v, want saved message from alice with trimmed text", name, msg)
		}
	}
	if ev, ok := elsewhere.next(t, 300*time.Millisecond); ok {
		t.Errorf("client in another room received %+v", ev)
	}

	history := s.history(t, "general", "")
	if len(history) != 1 || history[0].Text != "hello bob" || history[0].Username != "alice" {
		t.Fatalf("history = %+v, want alice's message", history)
	}
	if s.history(t, "random", "") == nil {
		t.Errorf("empty room history decoded as null, want []")
	}
}

func TestTypingEventsAreNotPersisted(t *testing.T) {
	s := newServer(t)
	alice := s.dial(t, "general", s.registerAndLogin(t, "alice"))
	bob := s.dial(t, "general", s.registerAndLogin(t, "bob"))
	settle()

	alice.send(t, map[string]string{"type": "typing"})
	if ev := bob.nextOfType(t, "typing"); ev.Username != "alice" {
		t.Errorf("typing event from %q, want alice", ev.Username)
	}
	if n := s.messageCount(t); n != 0 {
		t.Errorf("%d messages saved after a typing event, want 0", n)
	}
}

func TestGuestsCanReadButNotPost(t *testing.T) {
	s := newServer(t)
	alice := s.dial(t, "general", s.registerAndLogin(t, "alice"))
	guest := s.dial(t, "general", "")
	settle()

	guest.send(t, map[string]string{"type": "chat", "text": "guest message"})
	guest.send(t, map[string]string{"type": "typing"})
	settle()
	alice.send(t, map[string]string{"type": "chat", "text": "hello guest"})

	// Events arrive in order, so if the guest's got through they'd come first
	if ev, _ := alice.next(t, 2*time.Second); ev.Type != "chat" || decodeChat(t, ev).Text != "hello guest" {
		t.Errorf("member's first event = %+v, want only their own message", ev)
	}
	if msg := decodeChat(t, guest.nextOfType(t, "chat")); msg.Text != "hello guest" {
		t.Errorf("guest received %q, want hello guest", msg.Text)
	}
	if n := s.messageCount(t); n != 1 {
		t.Errorf("%d messages saved, want 1 (none from the guest)", n)
	}
}

func TestOversizedMessageIsRejectedWithoutDisconnecting(t *testing.T) {
	s := newServer(t)
	alice := s.dial(t, "general", s.registerAndLogin(t, "alice"))
	bob := s.dial(t, "general", s.registerAndLogin(t, "bob"))
	settle()

	alice.send(t, map[string]string{"type": "chat", "text": strings.Repeat("x", 1001)})
	if ev := alice.nextOfType(t, "error"); ev.Code != "message_too_long" {
		t.Errorf("error code = %q, want message_too_long", ev.Code)
	}
	if n := s.messageCount(t); n != 0 {
		t.Errorf("%d messages saved for a rejected message, want 0", n)
	}

	// The limit counts characters, not bytes, and the connection still works.
	// This must be the first thing bob sees: the rejected message never reached him.
	atLimit := strings.Repeat("é", 1000)
	alice.send(t, map[string]string{"type": "chat", "text": atLimit})
	if ev, _ := bob.next(t, 2*time.Second); ev.Type != "chat" || decodeChat(t, ev).Text != atLimit {
		t.Errorf("bob's first event = type %q, want the message at the limit, delivered intact", ev.Type)
	}
}

func TestEmptyChatMessagesAreIgnored(t *testing.T) {
	s := newServer(t)
	alice := s.dial(t, "general", s.registerAndLogin(t, "alice"))
	settle()

	alice.send(t, map[string]string{"type": "chat", "text": "   "})
	if ev, ok := alice.next(t, 300*time.Millisecond); ok {
		t.Errorf("received %+v for a blank message", ev)
	}
	if n := s.messageCount(t); n != 0 {
		t.Errorf("%d messages saved, want 0", n)
	}
}

func TestHistoryReturnsNewestMessagesInOrder(t *testing.T) {
	s := newServer(t)
	start := time.Now().UTC().Add(-time.Hour)
	for i := 1; i <= 60; i++ {
		store.AddMessage(models.Message{
			RoomID:    "general",
			Username:  "alice",
			Text:      fmt.Sprintf("msg %d", i),
			Timestamp: start.Add(time.Duration(i) * time.Second),
		})
	}

	tests := []struct {
		query     string
		wantFirst int
		wantLen   int
	}{
		{"", 11, 50},           // default limit
		{"?limit=5", 56, 5},    // custom limit
		{"?limit=0", 11, 50},   // invalid limits fall back to the default
		{"?limit=500", 11, 50}, // above the maximum of 200
		{"?limit=abc", 11, 50},
	}
	for _, tt := range tests {
		t.Run("query="+tt.query, func(t *testing.T) {
			got := s.history(t, "general", tt.query)
			if len(got) != tt.wantLen {
				t.Fatalf("got %d messages, want %d", len(got), tt.wantLen)
			}
			for i, m := range got {
				if want := fmt.Sprintf("msg %d", tt.wantFirst+i); m.Text != want {
					t.Fatalf("message %d = %q, want %q", i, m.Text, want)
				}
			}
		})
	}
}

func TestHistoryRejectsBadPaths(t *testing.T) {
	s := newServer(t)

	for _, path := range []string{"/api/rooms/general", "/api/rooms/general/other", "/api/rooms//messages"} {
		if status, _ := s.do(t, http.MethodGet, path, "", nil); status != http.StatusNotFound && status != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want 404 or 400", path, status)
		}
	}
	if status, _ := s.do(t, http.MethodPost, "/api/rooms/general/messages", "", nil); status != http.StatusMethodNotAllowed {
		t.Errorf("POST status = %d, want 405", status)
	}
}

func TestMessagesSurviveReconnect(t *testing.T) {
	s := newServer(t)
	token := s.registerAndLogin(t, "alice")

	first := s.dial(t, "general", token)
	settle()
	first.send(t, map[string]string{"type": "chat", "text": "before reconnect"})
	sent := decodeChat(t, first.nextOfType(t, "chat"))
	first.conn.Close()

	s.dial(t, "general", token)
	history := s.history(t, "general", "")
	if len(history) != 1 || history[0].ID != sent.ID {
		t.Errorf("history after reconnect = %+v, want the message sent earlier (id %d)", history, sent.ID)
	}
}
