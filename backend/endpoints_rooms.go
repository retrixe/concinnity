package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	nanoid "github.com/matoous/go-nanoid/v2"
)

type roomEndpointBody struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Target string `json:"target"`
}

func readRoomEndpointBody(r *http.Request, data *roomEndpointBody) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errorJson("Unable to read body!")
	}
	err = json.Unmarshal(body, data)
	if err != nil {
		return errorJson("Unable to read body!")
	} else if data.Type != "" && data.Type != "local_file" && data.Type != "remote_file" {
		return errorJson("Invalid room type!")
	} else if data.Type != "" && data.Target == "" {
		return errorJson("Target cannot be empty with room type '" + data.Type + "'!")
	}
	return ""
}

func CreateRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	if user, _ := IsAuthenticatedHTTP(w, r); user == nil {
		return
	}

	var body roomEndpointBody
	if err := readRoomEndpointBody(r, &body); err != "" {
		http.Error(w, err, http.StatusBadRequest)
		return
	}

	id := body.ID
	if id == "" {
		id = nanoid.Must(12)
	} else if res, _ := regexp.MatchString("^[a-zA-Z0-9_-]{0,24}$", id); !res {
		http.Error(w, errorJson("Invalid room ID!"), http.StatusBadRequest)
		return
	}

	result, err := insertRoomStmt.Exec(id, body.Type, body.Target)
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
		http.Error(w, errorJson("Room ID already exists!"), http.StatusConflict)
		return
	} else if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		http.Error(w, errorJson("Room ID already exists!"), http.StatusConflict)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte("{\"id\":\"" + id + "\"}"))
}

func GetRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	if user, _ := IsAuthenticatedHTTP(w, r); user == nil {
		return
	}

	room := Room{}
	err := findRoomStmt.QueryRow(r.PathValue("id")).Scan(
		&room.ID, &room.CreatedAt, &room.ModifiedAt, &room.Type, &room.Target,
		&room.Paused, &room.Speed, &room.Timestamp, &room.LastAction)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	room.Chat, err = FindChatMessagesByRoom(room.ID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	room.Subtitles, err = FindSubtitlesByRoom(room.ID)
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	json.NewEncoder(w).Encode(room)
}

func UpdateRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	if user, _ := IsAuthenticatedHTTP(w, r); user == nil {
		return
	}

	var body roomEndpointBody
	if err := readRoomEndpointBody(r, &body); err != "" {
		http.Error(w, err, http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	createdAt, modifiedAt, err := UpdateRoom(id, body.Type, body.Target)
	if err == sql.ErrNoRows {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Send message to all room members about the change
	if members, ok := roomMembers.Load(id); ok {
		members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
			write <- RoomInfoMessageOutgoing{
				Type: "room_info",
				Data: RoomInfoMessageOutgoingData{
					ID:         id,
					CreatedAt:  &createdAt,
					ModifiedAt: &modifiedAt,
					Type:       body.Type,
					Target:     body.Target,
				},
			}
			return true
		})
	}
	w.Write([]byte("{\"success\":true}"))
}

func GetRoomSubtitleEndpoint(w http.ResponseWriter, r *http.Request) {
	if user, _ := IsAuthenticatedHTTP(w, r); user == nil {
		return
	} else if r.URL.Query().Get("name") == "" {
		http.Error(w, errorJson("Name cannot be empty!"), http.StatusBadRequest)
		return
	}

	var subtitle string
	err := findSubtitleStmt.QueryRow(r.PathValue("id"), r.URL.Query().Get("name")).Scan(&subtitle)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Subtitles or room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	w.Write([]byte(subtitle))
}

func CreateRoomSubtitleEndpoint(w http.ResponseWriter, r *http.Request) {
	if user, _ := IsAuthenticatedHTTP(w, r); user == nil {
		return
	} else if r.URL.Query().Get("name") == "" {
		http.Error(w, errorJson("Name cannot be empty!"), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024*1024)) // 1 MB limit
	if err != nil || len(body) == 0 {
		http.Error(w, errorJson("Unable to read body!"), http.StatusBadRequest)
		return
	} else if len(body) == 1024*1024 {
		http.Error(w, errorJson("Body too large!"), http.StatusRequestEntityTooLarge)
		return
	}

	var result sql.Result
	if config.Database == "mysql" {
		result, err = insertSubtitleStmt.Exec(r.PathValue("id"), r.URL.Query().Get("name"), body, body)
	} else {
		result, err = insertSubtitleStmt.Exec(r.PathValue("id"), r.URL.Query().Get("name"), body)
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Class() == "23503" {
		http.Error(w, errorJson("Room does not exist!"), http.StatusNotFound)
		return
	} else if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
		http.Error(w, errorJson("Room does not exist!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err)
		return
	}

	// Send message to all room members about the change
	if members, ok := roomMembers.Load(r.PathValue("id")); ok {
		members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
			write <- SubtitleMessageOutgoing{Type: "subtitle", Data: []string{r.URL.Query().Get("name")}}
			return true
		})
	}

	w.Write([]byte("{\"success\":true}"))
}
