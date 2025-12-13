package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) AddMessage(msg models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO messages (room_id, username, text, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.ExecContext(ctx, query, msg.RoomID, msg.Username, msg.Text, msg.Timestamp)
	return err
}

func (s *PostgresStore) GetRecentMessages(roomID string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT room_id, username, text, timestamp
		FROM messages
		WHERE room_id = $1
		ORDER BY timestamp ASC
		LIMIT $2
	`
	rows, err := s.db.QueryContext(ctx, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.RoomID, &m.Username, &m.Text, &m.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
