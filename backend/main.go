package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

/*
Endpoints:
- GET /
- POST /api/login
- POST /api/logout
- GET/POST/PATCH/DELETE /api/room/:id
- WS /api/room/:id/connect
- POST /api/room/:id/join
*/

// TODO: use cookies to store token on front-end
func main() {
	// TODO: use environment variables
	connStr := "dbname=concinnity user=postgres host=localhost password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// TODO rooms - members, id, chat, timestamp, paused, lastActionTime, createdAt
	_, err = db.Exec(
		`CREATE TABLE users (
			username VARCHAR(16) UNIQUE,
			password VARCHAR(100),
			email TEXT UNIQUE,
			id UUID UNIQUE);`,
	)
	if err != nil {
		log.Panicln("Failed to create users table!", err)
	}
	_, err = db.Exec(
		`CREATE TABLE tokens (
			username VARCHAR(16) UNIQUE,
			token VARCHAR(128) UNIQUE,
			createdAt TIMESTAMPTZ,
			id UUID UNIQUE);`,
	)
	if err != nil {
		log.Panicln("Failed to create tokens table!", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write([]byte("{\"online\":true}"))
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
	})

	findUserStmt, err := db.Prepare(
		"SELECT username, password, email, id FROM users WHERE username = ? OR email = ? LIMIT 1;")
	if err != nil {
		log.Panicln(err)
	}
	insertTokenStmt, err := db.Prepare(
		"INSERT INTO tokens (username, token, createdAt, id) VALUES (?, ?, ?, ?);")
	if err != nil {
		log.Panicln(err)
	}
	http.HandleFunc("/api/login", LoginEndpoint(findUserStmt, insertTokenStmt))

	deleteTokenStmt, err := db.Prepare("DELETE FROM tokens WHERE token = ?;")
	if err != nil {
		log.Panicln(err)
	}
	http.HandleFunc("/api/logout", LogoutEndpoint(deleteTokenStmt))

	http.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else if r.Method == "POST" {
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else if r.Method == "PATCH" {
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else if r.Method == "DELETE" {
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
	})

	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Println("Listening to port " + port)
	http.ListenAndServe(":"+port, nil)
}
