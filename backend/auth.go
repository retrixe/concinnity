package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

func LoginEndpoint(findUserStmt *sql.Stmt, insertTokenStmt *sql.Stmt) func(
	w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Check the body for JSON containing username and password and return a token.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
				return
			}
			var data struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			err = json.Unmarshal(body, &data)
			if err != nil {
				http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
				return
			} else if data.Username == "" || data.Password == "" {
				http.Error(w, errorJson("No username or password provided!"), http.StatusBadRequest)
				return
			}
			var user User
			err = findUserStmt.QueryRow(data.Username, data.Username).Scan(
				&user.Username, &user.Password, &user.Email, &user.ID)
			if err != nil && errors.Is(err, sql.ErrNoRows) {
				http.Error(w, errorJson("No account with this username/email exists!"), http.StatusBadRequest)
				return
			} else if err != nil {
				http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
				return
			}
			tokenBytes := make([]byte, 64)
			_, _ = rand.Read(tokenBytes)
			token := hex.EncodeToString(tokenBytes)
			result, err := insertTokenStmt.Exec(user.Username, token, time.Now().UTC(), user.ID)
			if err != nil {
				http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
				return
			} else if rows, err := result.RowsAffected(); err != nil || rows == 1 {
				http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(struct {
				Token string `json:"token"`
			}{Token: token})
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
	}
}

func LogoutEndpoint(deleteTokenStmt *sql.Stmt) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "{\"error\":\"You are not authenticated to access this resource!\"}",
					http.StatusUnauthorized)
				return
			}
			result, err := deleteTokenStmt.Exec(token)
			if err != nil {
				http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
				return
			}
			rows, err := result.RowsAffected()
			if err != nil {
				http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
				return
			} else if rows == 0 {
				http.Error(w, "{\"error\":\"You are not authenticated to access this resource!\"}",
					http.StatusUnauthorized)
				return
			}
			w.Write([]byte("{\"success\":true}"))
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
	}
}
