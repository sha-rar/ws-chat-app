// backend/internal/store/store.go
package store

import (
	"log"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
)

type Store interface {
	AddMessage(msg models.Message) error
	GetRecentMessages(roomID string, limit int) ([]models.Message, error)
}

// defaultStore is the global store used by package-level helpers
var defaultStore Store

// InitDefaultStore sets the global store implementation
func InitDefaultStore(s Store) {
	defaultStore = s
}

// AddMessage is a convenience wrapper around defaultStore.AddMessage
func AddMessage(msg models.Message) {
	if defaultStore == nil {
		// no store configured; just log and return
		log.Printf("store.AddMessage called but defaultStore is nil")
		return
	}
	if err := defaultStore.AddMessage(msg); err != nil {
		log.Printf("store.AddMessage error: %v", err)
	}
}

// GetRecentMessages wraps defaultStore.GetRecentMessages
// If the store is nil or errors, it returns an empty slice
func GetRecentMessages(roomID string, limit int) []models.Message {
	if defaultStore == nil {
		log.Printf("store.GetRecentMessages called but defaultStore is nil")
		return []models.Message{}
	}
	msgs, err := defaultStore.GetRecentMessages(roomID, limit)
	if err != nil {
		log.Printf("store.GetRecentMessages error: %v", err)
		return []models.Message{}
	}
	return msgs
}
