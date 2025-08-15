package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type AuthMessageIncoming struct {
	Token     string `json:"token"`
	ClientID  string `json:"clientId"`
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

type TypingIndicatorMessageIncoming struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

type TypingIndicatorMessageOutgoing struct {
	Type      string    `json:"type"`
	UserID    uuid.UUID `json:"userId"`
	Timestamp int64     `json:"timestamp"`
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

type SubtitleMessageOutgoing struct {
	Type string   `json:"type"` // subtitle
	Data []string `json:"data"`
}

const (
	WsInternalAuthDisconnect = iota
	WsInternalClientReconnect
)

func JoinRoomEndpoint(w http.ResponseWriter, r *http.Request) {
	// Impl note: If target/type change, client should trash currently playing file, subs and reset state.
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
	} else if rooms, ok := userConns.Load(user.ID); ok && rooms.Size() >= 3 {
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
	subtitle, err := FindSubtitlesByRoom(room.ID)
	if err != nil {
		wsInternalError(c, err)
		return
	}

	// Send current room info, state, chat and subtitle
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
	err = wsjsonWriteWithTimeout(context.Background(), c,
		ChatMessageOutgoing{Type: "chat", Data: chat})
	if err != nil {
		wsError(c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}
	err = wsjsonWriteWithTimeout(context.Background(), c,
		SubtitleMessageOutgoing{Type: "subtitle", Data: subtitle})
	if err != nil {
		wsError(c, "Failed to write data!", websocket.StatusProtocolError)
		return
	}

	writeChannel := make(chan interface{}, 16)
	defer close(writeChannel)
	// Register user to room
	clientId := authMessage.ClientID
	if clientId == "" {
		clientId = rand.Text()
	}
	connId := RoomConnID{UserID: user.ID, ClientID: clientId}
	members, previousConnectionExisted :=
		RegisterConnection(room.ID, connId, authMessage.Token, writeChannel)
	defer UnregisterConnection(room.ID, connId, members, writeChannel)

	// Create write thread
	var silentlyDisconnect atomic.Bool
	go (func() {
		for msg := range writeChannel {
			switch msg {
			case WsInternalAuthDisconnect:
				wsError(c, "You are not authenticated to access this resource!", 4401)
				return
			case WsInternalClientReconnect:
				silentlyDisconnect.Store(true) // Don't notify other clients of a disconnect.
				wsError(c, "You reconnected from the same client instance!", 4401)
				return
			}
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
	// If not a reconnect, OR reconnect + no previous connection
	if !authMessage.Reconnect || (!previousConnectionExisted && authMessage.Reconnect) {
		chatMsg := ChatMessage{UserID: uuid.Nil}
		if authMessage.Reconnect {
			chatMsg.Message = user.ID.String() + " reconnected"
		} else {
			chatMsg.Message = user.ID.String() + " joined"
		}
		chatMsg.ID, chatMsg.Timestamp, err = InsertChatMessage(room.ID, nil, chatMsg.Message)
		if err != nil {
			wsInternalError(c, err)
			return
		}
		members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
			write <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
			return true
		})
	}

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
			msg := strings.TrimSpace(chatData.Data)
			if err != nil {
				wsError(c, "Invalid chat message!", websocket.StatusUnsupportedData)
				continue
			} else if len(msg) > 2000 || len(msg) == 0 {
				continue // Discard invalid length messages silently
			}

			// Update state in db and broadcast
			chatMsg := ChatMessage{UserID: user.ID, Message: msg}
			chatMsg.ID, chatMsg.Timestamp, err = InsertChatMessage(room.ID, &user.ID, chatMsg.Message)
			if err != nil {
				wsInternalError(c, err)
				return
			}
			members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
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
			var result sql.Result
			if config.Database == "mysql" {
				result, err = updateRoomStateStmt.Exec(
					playerStateData.Data.Paused, playerStateData.Data.Speed,
					playerStateData.Data.Timestamp, playerStateData.Data.LastAction,
					room.ID)
			} else {
				result, err = updateRoomStateStmt.Exec(room.ID,
					playerStateData.Data.Paused, playerStateData.Data.Speed,
					playerStateData.Data.Timestamp, playerStateData.Data.LastAction)
			}
			if err != nil {
				wsInternalError(c, err)
				return
			} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
				wsInternalError(c, err)
				return
			}
			members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
				if write != writeChannel { // Skip current session
					write <- playerStateData
				}
				return true
			})
		} else if msgData.Type == "typing" {
			var incoming TypingIndicatorMessageIncoming
			err = json.Unmarshal(data, &incoming)
			if err != nil {
				wsError(c, "Error while sending typing indicators!", websocket.StatusUnsupportedData)
				continue
			}
			outgoingData := TypingIndicatorMessageOutgoing{
				Type:      "typing",
				UserID:    user.ID,
				Timestamp: incoming.Timestamp,
			}
			members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
				if write != writeChannel { // Skip current session
					write <- outgoingData
				}
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
	if silentlyDisconnect.Load() {
		return
	}
	chatMsg := ChatMessage{UserID: uuid.Nil}
	if closeStatus == websocket.StatusNormalClosure || closeStatus == websocket.StatusGoingAway {
		chatMsg.Message = user.ID.String() + " left"
	} else {
		chatMsg.Message = user.ID.String() + " was disconnected"
	}
	chatMsg.ID, chatMsg.Timestamp, err = InsertChatMessage(room.ID, nil, chatMsg.Message)
	if err != nil {
		log.Println("Internal Server Error!", err)
		return
	}
	members.Range(func(connId RoomConnID, write chan<- interface{}) bool {
		write <- ChatMessageOutgoing{Type: "chat", Data: []ChatMessage{chatMsg}}
		return true
	})
}

func wsjsonWriteWithTimeout(ctx context.Context, c *websocket.Conn, v interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	return wsjson.Write(ctx, c, v)
}
