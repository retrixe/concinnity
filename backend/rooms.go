package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lib/pq"
	nanoid "github.com/matoous/go-nanoid/v2"
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
		Title string `json:"title"`
		Type  string `json:"type"`
		Extra string `json:"extra"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.Type != "localFile" && data.Type != "remoteFile" {
		http.Error(w, errorJson("Invalid room type!"), http.StatusBadRequest)
		return
	} else if data.Title == "" {
		http.Error(w, errorJson("Title cannot be empty!"), http.StatusBadRequest)
		return
	} else if data.Extra == "" {
		http.Error(w, errorJson("Extra data cannot be empty with '"+data.Type+"' type of room!"),
			http.StatusBadRequest)
		return
	}

	id := nanoid.Must(12)
	result, err := insertRoomStmt.Exec(id, data.Type, data.Title, data.Extra)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"id\":\"" + id + "\"}"))
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	token := Token{}
	if IsAuthenticated(w, r, &token) == nil {
		http.Error(w, errorJson("You are not authenticated!"), http.StatusForbidden)
		return
	}

	// Get the URL and extract the room ID from /api/rooms/:id
	id := r.URL.Path[len("/api/rooms"):]

	room := Room{}
	err := findRoomByIdStmt.QueryRow(id).Scan(
		&room.ID, &room.Type, &room.Title, &room.Extra,
		pq.Array(&room.Chat), &room.Paused, &room.Timestamp,
		&room.CreatedAt, &room.LastActionTime)
	if errors.Is(sql.ErrNoRows, err) {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	json.NewEncoder(w).Encode(room)
}
