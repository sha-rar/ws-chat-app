package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/db"
)

// SessionTTL is how long a login lasts before the user must log in again
const SessionTTL = 24 * time.Hour

// ErrInvalidToken covers bad signatures, expired tokens and logged-out sessions
var ErrInvalidToken = errors.New("invalid, expired or revoked token")

// CreateSession records a new login for the user and returns its ID and expiry
func CreateSession(ctx context.Context, userID int64) (string, time.Time, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generate session id: %w", err)
	}
	id := hex.EncodeToString(buf)
	expiresAt := time.Now().Add(SessionTTL)

	_, err := db.DB.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		id, userID, expiresAt,
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("insert session: %w", err)
	}
	return id, expiresAt, nil
}

// RevokeSession marks the session as logged out; it's a no-op if already revoked
func RevokeSession(ctx context.Context, id string) error {
	_, err := db.DB.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`,
		id,
	)
	return err
}

// Authenticate verifies the token's signature and checks that its session is still active
func Authenticate(ctx context.Context, tokenStr string) (*Claims, error) {
	claims, err := ParseToken(tokenStr)
	if err != nil || claims.ID == "" {
		return nil, ErrInvalidToken
	}

	var active bool
	err = db.DB.QueryRowContext(ctx, `
		SELECT revoked_at IS NULL AND expires_at > now()
		FROM sessions
		WHERE id = $1 AND user_id = $2
	`, claims.ID, claims.UserID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("check session: %w", err)
	}
	if !active {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
