package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/lib/pq"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/puzpuzpuz/xsync/v3"
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
	} else if data.Type != "" && data.Type != "localFile" && data.Type != "remoteFile" {
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
	// Send message to all room members about the change
	members, ok := roomMembers.Load(id)
	if ok {
		members.Range(func(key uuid.UUID, value chan<- interface{}) bool {
			value <- RoomInfoMessageOutgoing{
				Type: "room_info",
				Data: RoomInfoMessageOutgoingData{
					ID:         id,
					CreatedAt:  nil,
					ModifiedAt: nil,
					Type:       body.Type,
					Target:     body.Target,
				},
			}
			return true
		})
	}
	w.Write([]byte("{\"success\":true}"))
}

type AuthMessageIncoming struct {
	Token     string `json:"token"`
	Reconnect bool   `json:"reconnect"` // If this is a reconnect
}

type GenericMessage struct {
	Type string `json:"type"`
}

type ChatMessageIncoming struct {
	Type string `json:"type"` // chat
	Data string `json:"data"`
}

type PingPongMessageBi struct {
	Type      string `json:"type"` // ping if incoming, pong if outgoing
	Timestamp int    `json:"timestamp"`
}

type PlayerStateMessageBi struct {
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
	ID         string     `json:"id"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
	ModifiedAt *time.Time `json:"modifiedAt,omitempty"`
	Type       string     `json:"type"`
	Target     string     `json:"target"`
}

type ChatMessageOutgoing struct {
	Type string        `json:"type"` // chat
	Data []ChatMessage `json:"data"`
}

func JoinRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	// Impl note: If target/type change, client should trash currently playing file and reset state.

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"v0"}})
	if err != nil {
		http.Error(w, "Unable to upgrade connection!", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Wait for auth message
	var authMessage AuthMessageIncoming
	if err := wsjson.Read(ctx, c, &authMessage); err != nil {
		wsError(ctx, c, "Unable to read authentication message!", websocket.StatusProtocolError)
		return
	}
	user, _, err := IsAuthenticated(authMessage.Token)
	if errors.Is(err, ErrNotAuthenticated) {
		wsError(ctx, c, "You are not authenticated to access this resource!", 4401)
		return
	} else if err != nil {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	} else if rooms, ok := userRooms.Load(user.ID); ok && rooms.Load() >= 3 {
		wsError(ctx, c, "You are in too many rooms!", 4429)
		return
	}

	// Get room details, if not exists, boohoo
	room := Room{}
	err = findRoomStmt.QueryRow(r.PathValue("id")).Scan(
		&room.ID, &room.CreatedAt, &room.ModifiedAt,
		&room.Type, &room.Target, pq.Array(&room.Chat),
		&room.Paused, &room.Speed, &room.Timestamp, &room.LastAction)
	if errors.Is(err, sql.ErrNoRows) {
		wsError(ctx, c, "Room not found!", 4404)
		return
	} else if err != nil {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	}

	// Send current room info, state and chat
	err = wsjson.Write(ctx, c, RoomInfoMessageOutgoing{
		Type: "room_info",
		Data: RoomInfoMessageOutgoingData{
			ID:         room.ID,
			CreatedAt:  &room.CreatedAt,
			ModifiedAt: &room.ModifiedAt,
			Type:       room.Type,
			Target:     room.Target,
		},
	})
	if err != nil {
		wsError(ctx, c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}
	err = wsjson.Write(ctx, c, PlayerStateMessageBi{
		Type: "player_state",
		Data: PlayerStateMessageData{
			Paused:     room.Paused,
			Speed:      room.Speed,
			Timestamp:  room.Timestamp,
			LastAction: room.LastAction,
		},
	})
	if err != nil {
		wsError(ctx, c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}
	err = wsjson.Write(ctx, c, ChatMessageOutgoing{Type: "chat", Data: room.Chat})
	if err != nil {
		wsError(ctx, c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}

	writeChannel := make(chan interface{}, 16)
	// Register user to room
	members, _ := roomMembers.LoadOrStore(room.ID, xsync.NewMapOf[uuid.UUID, chan<- interface{}]())
	members.Store(user.ID, writeChannel)
	roomCounter, _ := userRooms.LoadOrStore(user.ID, atomic.Int32{})
	roomCounter.Add(1)

	// Create write thread
	go (func() {
		defer close(writeChannel)
		defer members.Delete(user.ID)
		defer (func() {
			if val := roomCounter.Add(-1); val == 0 {
				userRooms.Delete(user.ID)
			}
		})()
		for msg, ok := <-writeChannel; ok; {
			err := wsjson.Write(ctx, c, msg)
			if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) { // TODO correct?
				return
			} else if err != nil {
				wsError(ctx, c, "Failed to write data!", websocket.StatusProtocolError)
				return
			}
		}
	})()

	// Send chat message: user joined/reconnected
	chatMsg := ChatMessage{
		UserID:    uuid.Nil,
		Message:   user.ID.String() + " ",
		Timestamp: time.Now(),
	}
	if authMessage.Reconnect {
		chatMsg.Message += "reconnected"
	} else {
		chatMsg.Message += "joined"
	}
	result, err := insertChatMessageRoomStmt.Exec(room.ID, chatMsg)
	if err != nil {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	}
	members.Range(func(key uuid.UUID, value chan<- interface{}) bool {
		value <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
		return true
	})

	// Read all messages
	var closeStatus websocket.StatusCode = -1
	for {
		_, data, err := c.Read(ctx)
		closeStatus = websocket.CloseStatus(err)
		// TODO: Is this correct? What are the possible errors :/
		if closeStatus != -1 ||
			errors.Is(err, io.EOF) ||
			errors.Is(err, net.ErrClosed) ||
			errors.Is(err, context.Canceled) {
			break
		} else if err != nil {
			wsError(ctx, c, "Failed to read message!", websocket.StatusProtocolError)
			continue
		}

		// Parse message
		var msgData GenericMessage
		err = json.Unmarshal(data, &msgData)
		if err != nil {
			wsError(ctx, c, "Invalid message!", websocket.StatusUnsupportedData)
		} else if msgData.Type == "chat" {
			var chatData ChatMessageIncoming
			err = json.Unmarshal(data, &chatData)
			if err != nil {
				wsError(ctx, c, "Invalid chat message!", websocket.StatusUnsupportedData)
				continue
			}

			// Update state in db and broadcast
			result, err = insertChatMessageRoomStmt.Exec(room.ID, ChatMessage{
				UserID:    user.ID,
				Message:   chatData.Data,
				Timestamp: time.Now(),
			})
			if err != nil {
				wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
				return
			} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
				wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
				return
			}
			members.Range(func(key uuid.UUID, value chan<- interface{}) bool {
				if key == user.ID {
					return true // Skip current user
				}
				value <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
				return true
			})
		} else if msgData.Type == "player_state" {
			var playerStateData PlayerStateMessageBi
			err = json.Unmarshal(data, &playerStateData)
			if err != nil {
				wsError(ctx, c, "Invalid player state message!", websocket.StatusUnsupportedData)
				continue
			}

			// Update state in db and broadcast
			result, err = updateRoomStateStmt.Exec(room.ID,
				playerStateData.Data.Paused, playerStateData.Data.Speed,
				playerStateData.Data.Timestamp, playerStateData.Data.LastAction)
			if err != nil {
				wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
				return
			} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
				wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
				return
			}
			members.Range(func(key uuid.UUID, value chan<- interface{}) bool {
				if key == user.ID {
					return true // Skip current user
				}
				value <- playerStateData
				return true
			})
		} else if msgData.Type == "ping" {
			var pingData PingPongMessageBi
			err = json.Unmarshal(data, &pingData)
			if err != nil {
				wsError(ctx, c, "Invalid ping message!", websocket.StatusUnsupportedData)
				continue
			}
			err = wsjson.Write(ctx, c, PingPongMessageBi{Type: "pong", Timestamp: pingData.Timestamp})
			if err != nil {
				wsError(ctx, c, "Failed to write data!", websocket.StatusProtocolError)
			}
		} else {
			wsError(ctx, c, "Invalid message!", websocket.StatusUnsupportedData)
		}
	}

	// Notify other clients of the disconnect
	chatMsg = ChatMessage{
		UserID:    uuid.Nil,
		Message:   user.ID.String() + " disconnected",
		Timestamp: time.Now(),
	}
	if closeStatus != websocket.StatusNormalClosure && closeStatus != websocket.StatusGoingAway {
		chatMsg.Message += " unexpectedly"
	}
	result, err = insertChatMessageRoomStmt.Exec(room.ID, chatMsg)
	if err != nil {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		wsError(ctx, c, "Internal Server Error!", websocket.StatusInternalError)
		return
	}
	members.Range(func(key uuid.UUID, value chan<- interface{}) bool {
		value <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
		return true
	})
}
