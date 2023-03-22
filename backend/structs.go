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
	ID        []byte    `json:"id"`
}

type Room struct {
	ID             []byte    `json:"id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Extra          string    `json:"extra"`
	Chat           []string  `json:"chat"`
	Paused         bool      `json:"paused"`
	Timestamp      int       `json:"timestamp"`
	CreatedAt      time.Time `json:"createdAt"`
	LastActionTime time.Time `json:"lastActionTime"`
}
