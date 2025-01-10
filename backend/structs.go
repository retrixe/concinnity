package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	Verified  bool      `json:"verified"`
}

type Token struct {
	CreatedAt time.Time `json:"createdAt"`
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"userId"`
}

type Room struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Type       string    `json:"type"`
	Target     string    `json:"target"`

	Chat []ChatMessage `json:"chat,omitempty"`

	Paused     bool      `json:"paused"`
	Speed      float64   `json:"speed"`
	Timestamp  float64   `json:"timestamp"`
	LastAction time.Time `json:"lastAction"`
}

type ChatMessage struct {
	ID        int       `json:"id"`
	RoomID    string    `json:"roomId,omitempty"`
	UserID    uuid.UUID `json:"userId"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *ChatMessage) Scan(src interface{}) error {
	data, ok := src.([]byte)
	dataStr, okStr := src.(string)
	if !ok && !okStr {
		return errors.New("invalid type for chat message")
	} else if okStr {
		data = []byte(dataStr)
	}
	return json.Unmarshal(data, c)
}

func (c ChatMessage) Value() (driver.Value, error) {
	return json.Marshal(c)
}
