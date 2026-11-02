package websocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/auth"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/store"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize caps a raw frame; anything larger is treated as abuse and closes the connection
	maxMessageSize = 8192
	// MaxTextLength caps a chat message; longer ones are rejected with an error event
	MaxTextLength = 1000
)

// Client represents a single WebSocket connection
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	roomID    string
	userID    int64
	username  string
	sessionID string
	// guests can receive room events but not send any
	isGuest bool
}

// ServeWs upgrades the HTTP request to a WebSocket and registers a client
func ServeWs(hub *Hub, upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	// Auth: extract token from query; no token means a read-only guest
	var claims *auth.Claims
	tokenStr := r.URL.Query().Get("token")
	if tokenStr != "" {
		parsed, err := auth.Authenticate(r.Context(), tokenStr)
		if err != nil {
			if !errors.Is(err, auth.ErrInvalidToken) {
				log.Printf("websocket authenticate: %v", err)
			}
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		claims = parsed
	}

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		roomID = "general"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		roomID:  roomID,
		isGuest: claims == nil,
	}
	if claims != nil {
		client.userID = claims.UserID
		client.username = claims.Username
		client.sessionID = claims.ID
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket and forwards them to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			break
		}

		if c.isGuest {
			continue
		}

		messageBytes = bytes.TrimSpace(messageBytes)

		var incoming models.IncomingEvent
		if err := json.Unmarshal(messageBytes, &incoming); err != nil {
			log.Printf("invalid WS JSON: %v", err)
			continue
		}

		switch incoming.Type {
		case "chat":
			text := strings.TrimSpace(incoming.Text)
			if text == "" {
				continue
			}
			if utf8.RuneCountInString(text) > MaxTextLength {
				c.sendError("message_too_long", fmt.Sprintf("Messages can be at most %d characters.", MaxTextLength))
				continue
			}

			msg := models.Message{
				Username:  c.username,
				RoomID:    c.roomID,
				Text:      text,
				Timestamp: time.Now().UTC(),
			}

			// Store in DB
			store.AddMessage(msg)

			// Broadcast chat event
			out := models.OutgoingChatEvent{
				Type:    "chat",
				Message: msg,
			}
			outBytes, err := json.Marshal(out)
			if err != nil {
				log.Printf("failed to marshal chat event: %v", err)
				continue
			}
			c.hub.broadcast <- BroadcastMessage{
				RoomID: c.roomID,
				Data:   outBytes,
			}

		case "typing":
			// broadcast typing event, no DB write
			event := models.TypingEvent{
				Type:     "typing",
				Username: c.username,
				RoomID:   c.roomID,
				IsTyping: true,
			}
			outBytes, err := json.Marshal(event)
			if err != nil {
				log.Printf("failed to marshal typing event: %v", err)
				continue
			}
			c.hub.broadcast <- BroadcastMessage{
				RoomID: c.roomID,
				Data:   outBytes,
			}

		default:
			log.Printf("unknown WS event type: %s", incoming.Type)
			continue
		}
	}
}

// sendError tells this client (and no one else) that its event was rejected
func (c *Client) sendError(code, message string) {
	outBytes, err := json.Marshal(models.ErrorEvent{
		Type:    "error",
		Code:    code,
		Message: message,
	})
	if err != nil {
		log.Printf("failed to marshal error event: %v", err)
		return
	}
	c.hub.direct <- DirectMessage{Client: c, Data: outBytes}
}

// writePump writes messages from the hub to the WebSocket and sends pings
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel (logout or slow client)
				_ = c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(message)

			// Add queued messages to this frame
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
