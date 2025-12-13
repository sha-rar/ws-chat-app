package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx as database/sql driver
)

var DB *sql.DB

func Connect() (*sql.DB, error) {
	if DB != nil {
		return DB, nil
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// default
		dsn = "postgres://chatuser:chatpass@localhost:5432/chatapp?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	DB = db
	return DB, nil
}
