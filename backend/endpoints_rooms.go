package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
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
	var createdAt, modifiedAt time.Time
	err := updateRoomStmt.QueryRow(id, body.Type, body.Target).Scan(&createdAt, &modifiedAt)
	if err == sql.ErrNoRows {
		http.Error(w, errorJson("Room not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Send message to all room members about the change
	members, ok := roomMembers.Load(id)
	if ok {
		members.Range(func(write chan<- interface{}, userId uuid.UUID) bool {
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
	Speed      float64   `json:"speed"`
	Timestamp  float64   `json:"timestamp"`
	LastAction time.Time `json:"lastAction"`
}

type RoomInfoMessageOutgoing struct {
	Type string                      `json:"type"` // room_info
	Data RoomInfoMessageOutgoingData `json:"data"`
}

type RoomInfoMessageOutgoingData struct {
	ID         string     `json:"id"`
	CreatedAt  *time.Time `json:"createdAt"`
	ModifiedAt *time.Time `json:"modifiedAt"`
	Type       string     `json:"type"`
	Target     string     `json:"target"`
}

type ChatMessageOutgoing struct {
	Type string        `json:"type"` // chat
	Data []ChatMessage `json:"data"`
}

func JoinRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	// Impl note: If target/type change, client should trash currently playing file and reset state.
	// Impl note: Room info updates are currently only sent on join and when the target/type change.

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{"v0"},
		InsecureSkipVerify: true})
	if err != nil {
		return
	}

	// Wait for auth message
	var authMessage AuthMessageIncoming
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	err = wsjson.Read(ctx, c, &authMessage)
	cancel()
	if err != nil {
		wsError(c, "Unable to read authentication message!", websocket.StatusProtocolError)
		return
	}
	user, _, err := IsAuthenticated(authMessage.Token)
	if errors.Is(err, ErrNotAuthenticated) {
		wsError(c, "You are not authenticated to access this resource!", 4401)
		return
	} else if err != nil {
		wsInternalError(c, err)
		return
	} else if rooms, ok := userRooms.Load(user.ID); ok && rooms.Load() >= 3 {
		wsError(c, "You are in too many rooms!", 4429)
		return
	}

	// Get room details, if not exists, boohoo
	room := Room{}
	err = findRoomStmt.QueryRow(r.PathValue("id")).Scan(
		&room.ID, &room.CreatedAt, &room.ModifiedAt, &room.Type, &room.Target,
		&room.Paused, &room.Speed, &room.Timestamp, &room.LastAction)
	if errors.Is(err, sql.ErrNoRows) {
		wsError(c, "Room not found!", 4404)
		return
	} else if err != nil {
		wsInternalError(c, err)
		return
	}
	chat, err := FindChatMessagesByRoom(room.ID)
	if err != nil {
		wsInternalError(c, err)
		return
	}

	// Send current room info, state and chat
	err = wsjsonWriteWithTimeout(context.Background(), c, RoomInfoMessageOutgoing{
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
		wsError(c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}
	err = wsjsonWriteWithTimeout(context.Background(), c, PlayerStateMessageBi{
		Type: "player_state",
		Data: PlayerStateMessageData{
			Paused:     room.Paused,
			Speed:      room.Speed,
			Timestamp:  room.Timestamp,
			LastAction: room.LastAction,
		},
	})
	if err != nil {
		wsError(c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}
	err = wsjsonWriteWithTimeout(context.Background(), c, ChatMessageOutgoing{
		Type: "chat", Data: chat})
	if err != nil {
		wsError(c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}

	writeChannel := make(chan interface{}, 16)
	defer close(writeChannel)
	// Register user to room
	members, _ := roomMembers.LoadOrStore(room.ID, xsync.NewMapOf[chan<- interface{}, uuid.UUID]())
	members.Store(writeChannel, user.ID)
	defer members.Delete(writeChannel)
	roomCounter, _ := userRooms.LoadOrStore(user.ID, &atomic.Int32{})
	roomCounter.Add(1)
	defer userRooms.Compute(user.ID, func(value *atomic.Int32, loaded bool) (*atomic.Int32, bool) {
		if val := value.Add(-1); val == 0 {
			return value, true
		}
		return value, false
	})

	// Create write thread
	go (func() {
		for msg := range writeChannel {
			err := wsjsonWriteWithTimeout(context.Background(), c, msg)
			if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) { // TODO correct?
				return
			} else if err != nil {
				wsError(c, "Failed to write data!", websocket.StatusProtocolError)
				return
			}
		}
	})()

	// Send chat message: user joined/reconnected
	chatMsg := ChatMessage{UserID: uuid.Nil}
	if authMessage.Reconnect {
		chatMsg.Message = user.ID.String() + " reconnected"
	} else {
		chatMsg.Message = user.ID.String() + " joined"
	}
	err = insertChatMessageStmt.QueryRow(room.ID, nil, chatMsg.Message).Scan(
		&chatMsg.ID, &chatMsg.Timestamp)
	if err != nil {
		wsInternalError(c, err)
		return
	}
	members.Range(func(write chan<- interface{}, userId uuid.UUID) bool {
		write <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
		return true
	})

	// Read all messages
	var closeStatus websocket.StatusCode = -1
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		_, data, err := c.Read(ctx)
		cancel()
		closeStatus = websocket.CloseStatus(err)
		// TODO: Is this correct? What are the possible errors :/
		if closeStatus != -1 ||
			errors.Is(err, io.EOF) ||
			errors.Is(err, net.ErrClosed) ||
			errors.Is(err, context.Canceled) {
			break
		} else if err != nil {
			wsError(c, "Failed to read message!", websocket.StatusProtocolError)
			continue
		}

		// Parse message
		var msgData GenericMessage
		err = json.Unmarshal(data, &msgData)
		if err != nil {
			wsError(c, "Invalid message!", websocket.StatusUnsupportedData)
		} else if msgData.Type == "chat" {
			var chatData ChatMessageIncoming
			err = json.Unmarshal(data, &chatData)
			// Enforce 2000 char chat message limit
			if err != nil || len(chatData.Data) > 2000 {
				wsError(c, "Invalid chat message!", websocket.StatusUnsupportedData)
				continue
			}

			// Update state in db and broadcast
			chatMsg := ChatMessage{UserID: user.ID, Message: chatData.Data}
			err = insertChatMessageStmt.QueryRow(room.ID, user.ID, chatMsg.Message).Scan(
				&chatMsg.ID, &chatMsg.Timestamp)
			if err != nil {
				wsInternalError(c, err)
				return
			}
			members.Range(func(write chan<- interface{}, userId uuid.UUID) bool {
				write <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
				return true
			})
		} else if msgData.Type == "player_state" {
			var playerStateData PlayerStateMessageBi
			err = json.Unmarshal(data, &playerStateData)
			if err != nil {
				wsError(c, "Invalid player state message!", websocket.StatusUnsupportedData)
				continue
			}

			// Update state in db and broadcast
			result, err := updateRoomStateStmt.Exec(room.ID,
				playerStateData.Data.Paused, playerStateData.Data.Speed,
				playerStateData.Data.Timestamp, playerStateData.Data.LastAction)
			if err != nil {
				wsInternalError(c, err)
				return
			} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
				wsInternalError(c, err)
				return
			}
			members.Range(func(write chan<- interface{}, userId uuid.UUID) bool {
				if write == writeChannel {
					return true // Skip current session
				}
				write <- playerStateData
				return true
			})
		} else if msgData.Type == "ping" {
			var pingData PingPongMessageBi
			err = json.Unmarshal(data, &pingData)
			if err != nil {
				wsError(c, "Invalid ping message!", websocket.StatusUnsupportedData)
				continue
			}
			writeChannel <- PingPongMessageBi{Type: "pong", Timestamp: pingData.Timestamp}
		} else {
			wsError(c, "Invalid message!", websocket.StatusUnsupportedData)
		}
	}

	// Notify other clients of the disconnect
	chatMsg = ChatMessage{UserID: uuid.Nil}
	if closeStatus == websocket.StatusNormalClosure || closeStatus == websocket.StatusGoingAway {
		chatMsg.Message = user.ID.String() + " left"
	} else {
		chatMsg.Message = user.ID.String() + " was disconnected"
	}
	err = insertChatMessageStmt.QueryRow(room.ID, nil, chatMsg.Message).Scan(
		&chatMsg.ID, &chatMsg.Timestamp)
	if err != nil {
		log.Println("Internal Server Error!", err)
		return
	}
	members.Range(func(write chan<- interface{}, userId uuid.UUID) bool {
		write <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
		return true
	})
}

func wsjsonWriteWithTimeout(ctx context.Context, c *websocket.Conn, v interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	return wsjson.Write(ctx, c, v)
}
