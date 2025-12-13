package models

import "time"

type Message struct {
	Username  string    `json:"username"`
	RoomID    string    `json:"roomId"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
