// backend/internal/store/memory.go
package store

import (
	"sync"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
)

type InMemoryStore struct {
	mu         sync.RWMutex
	rooms      map[string][]models.Message
	maxPerRoom int
}

func NewInMemoryStore(maxPerRoom int) *InMemoryStore {
	return &InMemoryStore{
		rooms:      make(map[string][]models.Message),
		maxPerRoom: maxPerRoom,
	}
}

func (s *InMemoryStore) AddMessage(msg models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := append(s.rooms[msg.RoomID], msg)
	if len(msgs) > s.maxPerRoom {
		start := len(msgs) - s.maxPerRoom
		msgs = msgs[start:]
	}
	s.rooms[msg.RoomID] = msgs
	return nil
}

func (s *InMemoryStore) GetRecentMessages(roomID string, limit int) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := s.rooms[roomID]
	if limit > 0 && len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}

	out := make([]models.Message, len(msgs))
	copy(out, msgs)
	return out, nil
}
