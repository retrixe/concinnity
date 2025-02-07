package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrNotAuthenticated = errors.New("request not authenticated")

func IsAuthenticatedHTTP(w http.ResponseWriter, r *http.Request) (*User, *Token) {
	user, token, err := IsAuthenticated(GetTokenFromHTTP(r))
	if errors.Is(err, ErrNotAuthenticated) {
		http.Error(w, errorJson("You are not authenticated to access this resource!"),
			http.StatusUnauthorized)
	} else if err != nil {
		handleInternalServerError(w, err)
	}
	return user, token
}

func IsAuthenticated(token string) (*User, *Token, error) {
	if token == "" {
		return nil, nil, ErrNotAuthenticated
	}

	user := User{}
	var tokenCreatedAt time.Time
	err := findUserByTokenStmt.QueryRow(token).Scan(
		&user.Username,
		&user.Password,
		&user.Email,
		&user.ID,
		&user.CreatedAt,
		&user.Verified,
		&token,
		&tokenCreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotAuthenticated
	} else if err != nil {
		return nil, nil, err
	} else {
		return &user, &Token{CreatedAt: tokenCreatedAt, Token: token, UserID: user.ID}, nil
	}
}

func StatusEndpoint(w http.ResponseWriter, r *http.Request) {
	user, _, err := IsAuthenticated(GetTokenFromHTTP(r))
	if errors.Is(err, ErrNotAuthenticated) {
		w.Write([]byte("{\"online\":true,\"authenticated\":false}"))
	} else if err != nil {
		handleInternalServerError(w, err)
	} else {
		usernameJson, _ := json.Marshal(user.Username)
		userIdJson, _ := json.Marshal(user.ID)
		w.Write([]byte("{\"online\":true,\"authenticated\":true," +
			"\"username\":" + string(usernameJson) + ",\"userId\":" + string(userIdJson) + "}"))
	}
}

func GetUsernamesEndpoint(w http.ResponseWriter, r *http.Request) {
	_, token := IsAuthenticatedHTTP(w, r)
	if token == nil {
		return
	}
	requestedIds := r.URL.Query()["id"]
	if len(requestedIds) == 0 {
		http.Error(w, errorJson("No user IDs provided!"), http.StatusBadRequest)
		return
	}
	ids := make([]uuid.UUID, len(requestedIds))
	var err error
	for i, id := range requestedIds {
		ids[i], err = uuid.Parse(id)
		if err != nil {
			http.Error(w, errorJson("Invalid user ID(s) provided!"), http.StatusBadRequest)
			return
		}
	}

	rows, err := findUsernamesByIdStmt.Query(pq.Array(requestedIds))
	if err != nil {
		handleInternalServerError(w, err)
		return
	}

	defer rows.Close()
	usernames := make(map[string]string)
	for rows.Next() {
		var id uuid.UUID
		var username string
		if err = rows.Scan(&id, &username); err != nil {
			handleInternalServerError(w, err)
			return
		}
		usernames[id.String()] = username
	}
	if rows.Err() != nil {
		handleInternalServerError(w, err)
		return
	}

	json.NewEncoder(w).Encode(usernames)
}

func LoginEndpoint(w http.ResponseWriter, r *http.Request) {
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
	err = findUserByNameOrEmailStmt.QueryRow(data.Username, data.Username).Scan(
		&user.Username, &user.Password, &user.Email, &user.ID, &user.CreatedAt, &user.Verified)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("No account with this username/email exists!"), http.StatusUnauthorized)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	} else if !user.Verified {
		http.Error(w, errorJson("Your account is not verified yet!"), http.StatusForbidden)
		return
	} else if !ComparePassword(data.Password, user.Password) {
		http.Error(w, errorJson("Incorrect password!"), http.StatusUnauthorized)
		return
	}
	tokenBytes := make([]byte, 64)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	result, err := insertTokenStmt.Exec(token, time.Now().UTC(), user.ID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	// Add cookie to browser.
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   config.SecureCookies,
		MaxAge:   3600 * 24 * 31,
		SameSite: http.SameSiteStrictMode,
		Path:     config.BasePath,
	})
	json.NewEncoder(w).Encode(struct {
		Token    string `json:"token"`
		Username string `json:"username"`
	}{Token: token, Username: user.Username})
}

func LogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	token := GetTokenFromHTTP(r)
	if token == "" {
		http.Error(w, errorJson("You are not authenticated to access this resource!"),
			http.StatusUnauthorized)
		return
	}
	var userID uuid.UUID
	err := deleteTokenStmt.QueryRow(token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, errorJson("You are not authenticated to access this resource!"),
			http.StatusUnauthorized)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Disconnect existing sessions
	if conns, ok := userConns.Load(userID); ok {
		conns.Range(func(key chan<- interface{}, value string) bool {
			if value == token {
				key <- nil
			}
			return true
		})
	}
	// Delete cookie on browser.
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "null",
		HttpOnly: true,
		Secure:   config.SecureCookies,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
	})
	w.Write([]byte("{\"success\":true}"))
}

func RegisterEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check the body for JSON containing username, password and email, and return a token.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.Username == "" || data.Password == "" || data.Email == "" {
		http.Error(w, errorJson("No username, e-mail or password provided!"), http.StatusBadRequest)
		return
	} else if data.Username == "system" { // Reserve this name to use in chat.
		http.Error(w, errorJson("An account with this e-mail already exists!"), http.StatusConflict)
		return
	} else if res, _ := regexp.MatchString("^[a-z0-9_]{4,16}$", data.Username); !res {
		http.Error(w, errorJson("Username should be 4-16 characters long, and "+
			"contain lowercase alphanumeric characters or _ only!"), http.StatusBadRequest)
		return
	} else if res, _ := regexp.MatchString("^.{8,64}$", data.Password); !res {
		http.Error(w, errorJson("Your password must be between 8 and 64 characters long!"),
			http.StatusBadRequest)
		return
	} else if res, _ := regexp.MatchString("^\\S+@\\S+\\.\\S+$", data.Email); !res {
		http.Error(w, errorJson("Invalid e-mail entered!"), http.StatusBadRequest)
		return
	}
	// Check if an account with this username or email already exists.
	var u User
	err = findUserByEmailStmt.QueryRow(data.Email).Scan(
		&u.Username, &u.Password, &u.Email, &u.ID, &u.CreatedAt, &u.Verified)
	if err == nil {
		http.Error(w, errorJson("An account with this e-mail already exists!"), http.StatusConflict)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		handleInternalServerError(w, err)
		return
	}
	err = findUserByUsernameStmt.QueryRow(data.Username).Scan(
		&u.Username, &u.Password, &u.Email, &u.ID, &u.CreatedAt, &u.Verified)
	if err == nil {
		http.Error(w, errorJson("An account with this username already exists!"), http.StatusConflict)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		handleInternalServerError(w, err)
		return
	}
	// Create the account.
	hash := HashPassword(data.Password, GenerateSalt())
	uuid, err := uuid.NewV7()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	result, err := createUserStmt.Exec(data.Username, hash, data.Email, uuid, true)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	w.Write([]byte("{\"success\":true}"))
}
