package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/db"
	apphttp "github.com/sha-rar/ws-chat-app/dev/backend/internal/http"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to DB
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	// Initialise the default store with Postgres
	pgStore := store.NewPostgresStore(database)
	store.InitDefaultStore(pgStore)

	hub := websocket.NewHub()
	go hub.Run()

	router := apphttp.NewRouter(hub)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on http://localhost:%s\n", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
