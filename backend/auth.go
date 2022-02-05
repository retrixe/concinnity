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

func IsAuthenticated(w http.ResponseWriter, r *http.Request, t *Token) *User {
	token := r.Header.Get("token")
	if cookie, err := r.Cookie("token"); err == nil {
		token = cookie.Value
	}

	res, err := findUserByTokenStmt.Query(token)
	if err != nil {
		http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
		return nil
	} else if !res.Next() {
		http.Error(w, errorJson("You are not authenticated to access this resource!"),
			http.StatusUnauthorized)
		return nil
	} else {
		// TODO: Set properties on t if not nil.
		return &User{} // TODO
	}
}

func StatusEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && IsAuthenticated(w, r, nil) != nil {
		w.Write([]byte("{\"online\":true,\"authenticated\":true}"))
	} else if r.Method == "GET" {
		w.Write([]byte("{\"online\":true,\"authenticated\":false}"))
	} else {
		http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
	}
}

func LoginEndpoint(w http.ResponseWriter, r *http.Request) {
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
		err = findUserByNameEmailStmt.QueryRow(data.Username, data.Username).Scan(
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
		// Add cookie to browser.
		r.AddCookie(&http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Secure:   secureCookies,
			MaxAge:   3600 * 24 * 31,
			SameSite: http.SameSiteStrictMode,
		})
		json.NewEncoder(w).Encode(struct {
			Token string `json:"token"`
		}{Token: token})
	} else {
		http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
	}
}

func LogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		token := Token{}
		if IsAuthenticated(w, r, &token) == nil {
			return
		}
		result, err := deleteTokenStmt.Exec(token.Token)
		if err != nil {
			http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
			return
		}
		rows, err := result.RowsAffected()
		if err != nil {
			http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
			return
		} else if rows == 0 {
			http.Error(w, errorJson("You are not authenticated to access this resource!"),
				http.StatusUnauthorized)
			return
		}
		// Delete cookie on browser.
		r.AddCookie(&http.Cookie{
			Name:     "token",
			Value:    "null",
			HttpOnly: true,
			Secure:   secureCookies,
			MaxAge:   -1,
			SameSite: http.SameSiteStrictMode,
		})
		w.Write([]byte("{\"success\":true}"))
	} else {
		http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
	}
}
