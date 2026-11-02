package models

import "time"

type Message struct {
	// ID is 0 if the message couldn't be saved
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	RoomID    string    `json:"roomId"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
