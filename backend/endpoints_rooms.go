package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/lib/pq"
	nanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/net/websocket"
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
	} else if data.Type != "localFile" && data.Type != "remoteFile" {
		return errorJson("Invalid room type!")
	} else if data.Target == "" {
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
	} else if res, _ := regexp.MatchString("^[a-zA-Z0-9_-]{24}$", id); !res {
		http.Error(w, errorJson("Invalid room ID!"), http.StatusBadRequest)
		return
	}

	result, err := insertRoomStmt.Exec(id, body.Type, body.Target)
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
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
	result, err := updateRoomStmt.Exec(id, body.Type, body.Target)
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows != 1 {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	}
	w.Write([]byte("{\"success\":true}"))
}

func JoinRoomEndpointHandshake(config *websocket.Config, req *http.Request) error {
	config.Protocol = []string{"v0"}
	return nil
}

type AuthMessageIncoming struct {
	Token string `json:"token"`
}

type ErrorMessageOutgoing struct {
	Error string `json:"error"`
}

type GenericMessage struct {
	Type string `json:"type"`
}

type ChatMessageIncoming struct {
	Type string `json:"type"` // chat
	Data string `json:"data"`
}

type PlayerStateMessageIncoming struct {
	Type string                 `json:"type"` // player_state
	Data PlayerStateMessageData `json:"data"`
}

type PlayerStateMessageOutgoing struct {
	Type string                 `json:"type"` // player_state
	Data PlayerStateMessageData `json:"data"`
}

type PlayerStateMessageData struct {
	Paused     bool      `json:"paused"`
	Speed      int       `json:"speed"`
	Timestamp  int       `json:"timestamp"`
	LastAction time.Time `json:"lastAction"`
}

type RoomInfoMessageOutgoing struct {
	Type string                      `json:"type"` // room_info
	Data RoomInfoMessageOutgoingData `json:"data"`
}

type RoomInfoMessageOutgoingData struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Type       string    `json:"type"`
	Target     string    `json:"target"`
}

type ChatMessageOutgoing struct {
	Type string        `json:"type"` // chat
	Data []ChatMessage `json:"data"`
}

func JoinRoomEndpoint(ws *websocket.Conn) {
	// Impl note: If target/type change, client should trash currently playing file and reset state.

	// Wait for auth message
	ws.SetDeadline(time.Now().Add(30 * time.Second))
	var data AuthMessageIncoming
	if err := websocket.JSON.Receive(ws, &data); err != nil {
		wsError(ws, "Unable to read message!", 4400)
		return
	}
	user, _, err := IsAuthenticated(data.Token)
	if errors.Is(err, ErrNotAuthenticated) {
		wsError(ws, "You are not authenticated to access this resource!", 4401)
		return
	} else if err != nil {
		wsError(ws, "Internal Server Error!", 4500)
		return
	} else if rooms, ok := userRooms.Load(user.ID); ok && rooms.Load() >= 3 {
		wsError(ws, "You are in too many rooms!", 4429)
		return
	}

	// Get room details, if not exists, boohoo
	room := Room{}
	err = findRoomStmt.QueryRow(ws.Request().PathValue("id")).Scan(
		&room.ID, &room.CreatedAt, &room.ModifiedAt,
		&room.Type, &room.Target, pq.Array(&room.Chat),
		&room.Paused, &room.Speed, &room.Timestamp, &room.LastAction)
	if errors.Is(err, sql.ErrNoRows) {
		wsError(ws, "Room not found!", 4404)
		return
	} else if err != nil {
		wsError(ws, "Internal Server Error!", 4500)
		return
	}

	// Send current room info, state and chat
	err = websocket.JSON.Send(ws, RoomInfoMessageOutgoing{
		Type: "room_info",
		Data: RoomInfoMessageOutgoingData{
			ID:         room.ID,
			CreatedAt:  room.CreatedAt,
			ModifiedAt: room.ModifiedAt,
			Type:       room.Type,
			Target:     room.Target,
		},
	})
	if err != nil {
		wsError(ws, "Internal Server Error!", 4500)
		return
	}
	err = websocket.JSON.Send(ws, PlayerStateMessageOutgoing{
		Type: "player_state",
		Data: PlayerStateMessageData{
			Paused:     room.Paused,
			Speed:      room.Speed,
			Timestamp:  room.Timestamp,
			LastAction: room.LastAction,
		},
	})
	if err != nil {
		wsError(ws, "Internal Server Error!", 4500)
		return
	}
	err = websocket.JSON.Send(ws, ChatMessageOutgoing{
		Type: "chat",
		Data: room.Chat,
	})
	if err != nil {
		_ = websocket.JSON.Send(ws, ErrorMessageOutgoing{Error: "Internal Server Error!"})
		_ = ws.WriteClose(4500)
		return
	}

	// FIXME - Bump modifiedAt timestamp of room and add user to members
	// FIXME - User sends chat messages and state
	// FIXME - User receives chat messages, state changes and room info changes
	// FIXME - Add chat message on disconnect: user disconnected (abruptly)
}
