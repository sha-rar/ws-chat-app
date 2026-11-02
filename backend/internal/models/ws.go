package models

type IncomingEvent struct {
	Type string `json:"type"`           // "chat" or "typing"
	Text string `json:"text,omitempty"` // only for "chat"
}

type OutgoingChatEvent struct {
	Type    string  `json:"type"` // "chat"
	Message Message `json:"message"`
}

// ErrorEvent is sent only to the client whose event was rejected
type ErrorEvent struct {
	Type    string `json:"type"` // "error"
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TypingEvent struct {
	Type     string `json:"type"` // "typing"
	Username string `json:"username"`
	RoomID   string `json:"roomId"`
	IsTyping bool   `json:"isTyping"`
}
