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
	"strconv"
	"strings"
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
		emailJson, _ := json.Marshal(user.Email)
		w.Write([]byte("{\"online\":true,\"authenticated\":true," +
			"\"username\":" + string(usernameJson) +
			",\"userId\":" + string(userIdJson) +
			",\"email\":" + string(emailJson) + "}"))
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

	var rows *sql.Rows
	if config.Database == "mysql" {
		placeholders := strings.Repeat("?,", len(requestedIds))
		placeholders = placeholders[:len(placeholders)-1]
		mysqlArr := make([]interface{}, len(requestedIds))
		for i, id := range requestedIds {
			mysqlArr[i] = id
		}
		rows, err = prepareQuery(
			"SELECT id, username FROM users WHERE id IN (" + placeholders + ");").Query(mysqlArr...)
	} else {
		rows, err = findUsernamesByIdStmt.Query(pq.Array(requestedIds))
	}
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
	if errors.Is(err, sql.ErrNoRows) {
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
	if errors.Is(err, sql.ErrNoRows) {
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
				key <- WsInternalAuthDisconnect
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
	verified := true
	result, err := createUserStmt.Exec(data.Username, hash, data.Email, uuid, verified)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	w.Write([]byte("{\"success\":true,\"verified\":" + strconv.FormatBool(verified) + "}"))
}

func ForgotPasswordEndpoint(w http.ResponseWriter, r *http.Request) {
	if !IsEmailConfigured() || config.FrontendURL == "" {
		http.Error(w, errorJson("This functionality is unavailable on this Concinnity instance."),
			http.StatusNotImplemented)
		return
	}
	usernameEmail := r.URL.Query().Get("user")
	if usernameEmail == "" {
		http.Error(w, errorJson("No username or email provided!"), http.StatusBadRequest)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	defer tx.Rollback()
	// Get user info from the database.
	var user User
	err = tx.Stmt(findUserByNameOrEmailStmt).QueryRow(usernameEmail, usernameEmail).Scan(
		&user.Username, &user.Password, &user.Email, &user.ID, &user.CreatedAt, &user.Verified)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("No account with this username/email exists!"), http.StatusUnauthorized)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Check if a password reset token was requested for this user in the last 2 minutes.
	var lastToken PasswordResetToken
	err = tx.Stmt(findRecentPasswordResetTokensStmt).QueryRow(user.ID).Scan(
		&lastToken.ID, &lastToken.UserID, &lastToken.CreatedAt)
	if err == nil {
		http.Error(w, errorJson("A password reset token was already requested for this user in the last 2 minutes!"),
			http.StatusTooManyRequests)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		handleInternalServerError(w, err)
		return
	}
	// Insert a password reset token into the database.
	var token PasswordResetToken
	err = tx.Stmt(insertPasswordResetTokenStmt).QueryRow(user.ID).Scan(
		&token.ID, &token.UserID, &token.CreatedAt)
	if err != nil {
		handleInternalServerError(w, err) // An account was already confirmed to exist with this email.
		return
	}
	err = tx.Commit()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Send the password reset email.
	err = SendHTMLEmail(user.Email, "Password Reset Request for Concinnity",
		"<p>"+
			"Hello,<br>\n<br>\n"+
			"We received a request to reset your password. If you did not make this request, "+
			"please ignore this email.<br>\n<br>\n"+
			"To reset your password, please click the link below:<br>\n<br>\n"+
			"<a href=\""+config.FrontendURL+"/reset-password/"+token.ID.String()+"\">"+
			config.FrontendURL+"/reset-password/"+token.ID.String()+
			"</a><br>\n<br>\n"+
			"As a security measure, this link will expire in 10 minutes."+
			"</p>")
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"success\":true}"))
}

func ForgotPasswordTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, errorJson("No password reset token provided!"), http.StatusBadRequest)
		return
	} else if uuid.Validate(token) != nil {
		http.Error(w, errorJson("Invalid password reset token!"), http.StatusBadRequest)
		return
	}
	var response struct {
		UserID    string    `json:"userId"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"createdAt"`
	}
	err := findUserByPasswordResetTokenStmt.QueryRow(token).Scan(
		&response.UserID, &response.Username, &response.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Invalid password reset token!"), http.StatusBadRequest)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	} else if response.CreatedAt.Add(10 * time.Minute).Before(time.Now().UTC()) {
		http.Error(w, errorJson("This password reset token has expired!"), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func ResetPasswordEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check the body for JSON containing token and password.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}
	var data struct {
		Token    uuid.UUID `json:"token"`
		Password string    `json:"password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.Password == "" {
		http.Error(w, errorJson("No password provided!"), http.StatusBadRequest)
		return
	} else if res, _ := regexp.MatchString("^.{8,64}$", data.Password); !res {
		http.Error(w, errorJson("Your password must be between 8 and 64 characters long!"),
			http.StatusBadRequest)
		return
	}
	// Delete the token and update the user's password.
	tx, err := db.Begin()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	defer tx.Rollback()
	hashedPassword := HashPassword(data.Password, GenerateSalt())
	var token PasswordResetToken
	err = tx.Stmt(deletePasswordResetTokenStmt).QueryRow(data.Token).Scan(
		&token.UserID, &token.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Invalid password reset token!"), http.StatusBadRequest)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	} else if token.CreatedAt.Add(10 * time.Minute).Before(time.Now().UTC()) {
		err = tx.Commit() // Delete the token to prevent reuse.
		if err != nil {
			handleInternalServerError(w, err)
			return
		}
		http.Error(w, errorJson("This password reset token has expired!"), http.StatusBadRequest)
		return
	}
	result, err := tx.Stmt(updateUserPasswordStmt).Exec(hashedPassword, token.UserID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	err = tx.Commit()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"success\":true}"))
}

func ChangePasswordEndpoint(w http.ResponseWriter, r *http.Request) {
	user, token := IsAuthenticatedHTTP(w, r)
	if token == nil {
		return
	}
	// Check the body for JSON containing passwords.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}
	var data struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.CurrentPassword == "" {
		http.Error(w, errorJson("No current password provided!"), http.StatusBadRequest)
		return
	} else if data.NewPassword == "" {
		http.Error(w, errorJson("No new password provided!"), http.StatusBadRequest)
		return
	} else if !ComparePassword(data.CurrentPassword, user.Password) {
		http.Error(w, errorJson("Incorrect current password!"), http.StatusUnauthorized)
		return
	} else if res, _ := regexp.MatchString("^.{8,64}$", data.NewPassword); !res {
		http.Error(w, errorJson("Your password must be between 8 and 64 characters long!"),
			http.StatusBadRequest)
		return
	}
	hashedPassword := HashPassword(data.NewPassword, GenerateSalt())
	result, err := updateUserPasswordStmt.Exec(hashedPassword, token.UserID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	w.Write([]byte("{\"success\":true}"))
}

func DeleteAccountEndpoint(w http.ResponseWriter, r *http.Request) {
	user, token := IsAuthenticatedHTTP(w, r)
	if token == nil {
		return
	}
	// Check the body for JSON containing the user's password.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}
	var data struct {
		CurrentPassword string `json:"currentPassword"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.CurrentPassword == "" {
		http.Error(w, errorJson("No current password provided!"), http.StatusBadRequest)
		return
	} else if !ComparePassword(data.CurrentPassword, user.Password) {
		http.Error(w, errorJson("Incorrect current password!"), http.StatusUnauthorized)
		return
	}
	result, err := deleteUserStmt.Exec(token.UserID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	w.Write([]byte("{\"success\":true}"))
}
