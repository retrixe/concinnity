package main

import "time"

type User struct {
	ID        []byte    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	Verified  bool      `json:"verified"`
}

type Token struct {
	CreatedAt time.Time `json:"createdAt"`
	Token     string    `json:"token"`
	UserID    []byte    `json:"userId"`
}

type Room struct {
	ID         []byte    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Type       string    `json:"type"`
	Target     string    `json:"target"`
	Chat       []string  `json:"chat"`

	Paused     bool      `json:"paused"`
	Speed      int       `json:"speed"`
	Timestamp  int       `json:"timestamp"`
	LastAction time.Time `json:"lastAction"`
}
