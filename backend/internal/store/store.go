// backend/internal/store/store.go
package store

import (
	"log"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
)

type Store interface {
	// AddMessage saves msg and returns it with its assigned ID
	AddMessage(msg models.Message) (models.Message, error)
	GetRecentMessages(roomID string, limit int) ([]models.Message, error)
}

// defaultStore is the global store used by package-level helpers
var defaultStore Store

// InitDefaultStore sets the global store implementation
func InitDefaultStore(s Store) {
	defaultStore = s
}

// AddMessage is a convenience wrapper around defaultStore.AddMessage
// If the store is nil or errors, it returns msg unchanged (with ID 0)
func AddMessage(msg models.Message) models.Message {
	if defaultStore == nil {
		// no store configured; just log and return
		log.Printf("store.AddMessage called but defaultStore is nil")
		return msg
	}
	saved, err := defaultStore.AddMessage(msg)
	if err != nil {
		log.Printf("store.AddMessage error: %v", err)
		return msg
	}
	return saved
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
