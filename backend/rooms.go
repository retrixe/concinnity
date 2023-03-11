package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	// Check the body for JSON containing username and password and return a token.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}

	token := Token{}
	var user *User = IsAuthenticated(w, r, &token)
	if user == nil {
		http.Error(w, errorJson("You are not authenticated!"),
			http.StatusForbidden)
		return
	}

	var data struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		FileName string `json:"fileName"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	}

	// TODO: Use nanoid for IDs
	uuid := uuid.New()
	result, err := insertRoomStmt.Exec(
		uuid.String(), data.Type, data.Title, data.FileName, pq.Array([]string{user.Username}))
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"id\":\"" + uuid.String() + "\"}"))
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	token := Token{}
	if IsAuthenticated(w, r, &token) == nil {
		http.Error(w, errorJson("You are not authenticated!"),
			http.StatusForbidden)
		return
	}

	// Get the URL and extract the room ID from /api/rooms/:id
	id := r.URL.Path[len("/api/rooms"):]

	room := Room{}
	err := findRoomByIdStmt.QueryRow(id).Scan(
		&room.ID, &room.Type, &room.Title, &room.Extra,
		pq.Array(&room.Chat), pq.Array(&room.Members), &room.Paused, &room.Timestamp,
		&room.CreatedAt, &room.LastActionTime)
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	json.NewEncoder(w).Encode(room)
}
