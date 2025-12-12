package main

import (
	"log"
	"net/http"
	"os"
	"time"

	apphttp "github.com/sha-rar/ws-chat-app/dev/backend/internal/http"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create the WebSocket hub to handle all clients + broadcasts
	hub := websocket.NewHub()
	go hub.Run()

	// Build the HTTP router and inject the hub
	router := apphttp.NewRouter(hub)

	// HTTP server config
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
