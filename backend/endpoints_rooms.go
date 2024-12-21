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

func CreateRoomEndpoint(w http.ResponseWriter, r *http.Request) {
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
		Type   string `json:"type"`
		Target string `json:"target"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if data.Type != "localFile" && data.Type != "remoteFile" {
		http.Error(w, errorJson("Invalid room type!"), http.StatusBadRequest)
		return
	} else if data.Target == "" {
		http.Error(w, errorJson("Target cannot be empty with room type '"+data.Type+"'!"),
			http.StatusBadRequest)
		return
	}

	id := nanoid.Must(12)
	result, err := insertRoomStmt.Exec(id, data.Type, data.Target)
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"id\":\"" + id + "\"}"))
}

func GetRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	token := Token{}
	if IsAuthenticated(w, r, &token) == nil {
		http.Error(w, errorJson("You are not authenticated!"), http.StatusForbidden)
		return
	}

	room := Room{}
	err := findRoomByIdStmt.QueryRow(r.PathValue("id")).Scan(
		&room.ID, &room.CreatedAt, &room.ModifiedAt,
		&room.Type, &room.Target, pq.Array(&room.Chat),
		&room.Paused, &room.Speed, &room.Timestamp, &room.LastAction)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	json.NewEncoder(w).Encode(room)
}

func JoinRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}
