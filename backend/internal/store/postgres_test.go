package store_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/testutil"
)

// addMessages saves n messages to room, one second apart, ending now
func addMessages(t *testing.T, s store.Store, room string, n int) []models.Message {
	t.Helper()
	start := time.Now().UTC().Add(-time.Duration(n) * time.Second)
	saved := make([]models.Message, 0, n)
	for i := 1; i <= n; i++ {
		msg, err := s.AddMessage(models.Message{
			RoomID:    room,
			Username:  "alice",
			Text:      fmt.Sprintf("msg %d", i),
			Timestamp: start.Add(time.Duration(i) * time.Second),
		})
		if err != nil {
			t.Fatalf("AddMessage: %v", err)
		}
		saved = append(saved, msg)
	}
	return saved
}

func TestPostgresAddMessagePersists(t *testing.T) {
	conn := testutil.NewDB(t)
	s := store.NewPostgresStore(conn)

	saved := addMessages(t, s, "general", 2)
	if saved[0].ID == 0 || saved[1].ID <= saved[0].ID {
		t.Fatalf("IDs = %d, %d; want increasing non-zero IDs", saved[0].ID, saved[1].ID)
	}

	var room, username, text string
	err := conn.QueryRow(`SELECT room_id, username, text FROM messages WHERE id = $1`, saved[0].ID).
		Scan(&room, &username, &text)
	if err != nil {
		t.Fatalf("query saved message: %v", err)
	}
	if room != "general" || username != "alice" || text != "msg 1" {
		t.Errorf("row = (%q, %q, %q), want (general, alice, msg 1)", room, username, text)
	}
}

func TestPostgresGetRecentMessagesReturnsNewestInOrder(t *testing.T) {
	s := store.NewPostgresStore(testutil.NewDB(t))
	saved := addMessages(t, s, "general", 60)

	got, err := s.GetRecentMessages("general", 50)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 50 {
		t.Fatalf("got %d messages, want 50", len(got))
	}
	// The oldest 10 are dropped, and the rest come back oldest-first
	for i, m := range got {
		want := saved[i+10]
		if m.ID != want.ID || m.Text != want.Text {
			t.Fatalf("message %d = %d %q, want %d %q", i, m.ID, m.Text, want.ID, want.Text)
		}
	}
}

func TestPostgresGetRecentMessagesBreaksTimestampTiesByID(t *testing.T) {
	s := store.NewPostgresStore(testutil.NewDB(t))
	ts := time.Now().UTC()
	var ids []int64
	for _, text := range []string{"first", "second", "third"} {
		msg, err := s.AddMessage(models.Message{RoomID: "general", Username: "alice", Text: text, Timestamp: ts})
		if err != nil {
			t.Fatalf("AddMessage: %v", err)
		}
		ids = append(ids, msg.ID)
	}

	got, err := s.GetRecentMessages("general", 2)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 2 || got[0].ID != ids[1] || got[1].ID != ids[2] {
		t.Errorf("got %+v, want the last two (second, third) in order", got)
	}
}

func TestPostgresGetRecentMessagesIsolatesRooms(t *testing.T) {
	s := store.NewPostgresStore(testutil.NewDB(t))
	addMessages(t, s, "general", 3)
	addMessages(t, s, "random", 2)

	got, err := s.GetRecentMessages("random", 50)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d messages from random, want 2", len(got))
	}
	for _, m := range got {
		if m.RoomID != "random" {
			t.Errorf("got message from room %q in random's history", m.RoomID)
		}
	}
}

func TestPostgresGetRecentMessagesEmptyRoom(t *testing.T) {
	s := store.NewPostgresStore(testutil.NewDB(t))

	got, err := s.GetRecentMessages("nobody-here", 50)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want an empty, non-nil slice", got)
	}
}

func TestPostgresGetRecentMessagesDefaultsLimit(t *testing.T) {
	s := store.NewPostgresStore(testutil.NewDB(t))
	addMessages(t, s, "general", 55)

	got, err := s.GetRecentMessages("general", 0)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 50 {
		t.Errorf("got %d messages with limit 0, want the default of 50", len(got))
	}
}
