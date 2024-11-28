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
	ID        []byte    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Target    string    `json:"target"`
	Chat      []string  `json:"chat"`
	Members   [][]byte  `json:"members"`

	Paused     bool      `json:"paused"`
	Timestamp  int       `json:"timestamp"`
	LastAction time.Time `json:"lastAction"`
}
