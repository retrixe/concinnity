package main

import "time"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	ID       []byte `json:"id"`
}

type Token struct {
	CreatedAt time.Time `json:"createdAt"`
	Token     string    `json:"token"`
	ID        []byte    `json:"id"`
}

/* type Room struct { // TODO
	ID        []byte    `json:id`
	Timestamp time.Time `json:"timestamp"`
} */
