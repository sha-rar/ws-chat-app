// Package testutil provides helpers for tests that need a real Postgres database.
package testutil

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sha-rar/ws-chat-app/dev/backend/config"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/db"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
)

// NewDB returns a connection to a fresh, fully migrated schema and installs it as
// db.DB and the default store. The schema is dropped when the test finishes.
//
// Tests using it are skipped unless TEST_DATABASE_URL points at a Postgres database,
// e.g. postgres://chatuser:chatpass@localhost:5432/chatapp?sslmode=disable.
// It can be set in the environment or in the project's .env file.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()

	if _, err := config.LoadDotEnv(); err != nil {
		t.Fatalf("load .env: %v", err)
	}
	baseURL := os.Getenv("TEST_DATABASE_URL")
	if baseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run database tests")
	}

	admin, err := sql.Open("pgx", baseURL)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("generate schema name: %v", err)
	}
	schema := "test_" + hex.EncodeToString(suffix)
	if _, err := admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	// Point every connection in the pool at the new schema
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()

	conn, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("open test connection: %v", err)
	}
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	db.DB = conn
	store.InitDefaultStore(store.NewPostgresStore(conn))

	t.Cleanup(func() {
		db.DB = nil
		store.InitDefaultStore(nil)
		conn.Close()
		if _, err := admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Errorf("drop schema: %v", err)
		}
		admin.Close()
	})

	return conn
}
