package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/config"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/auth"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/db"
	apphttp "github.com/sha-rar/ws-chat-app/dev/backend/internal/http"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

func main() {
	envFile, err := config.LoadDotEnv()
	if err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}
	if envFile != "" {
		log.Printf("loaded settings from %s", envFile)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	auth.SetSecret(cfg.JWTSecret)

	// Connect to DB
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Initialise the default store with Postgres
	pgStore := store.NewPostgresStore(database)
	store.InitDefaultStore(pgStore)

	hub := websocket.NewHub()
	go hub.Run()

	router := apphttp.NewRouter(cfg, hub)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on http://localhost:%s (%s)\n", cfg.Port, cfg.Env)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
