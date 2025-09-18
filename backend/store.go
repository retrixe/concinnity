package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
)

type RoomConnID struct {
	UserID   uuid.UUID
	ClientID string
}

type UserConnInfo struct {
	RoomID string
	Token  string
}

type RoomMembers = *xsync.MapOf[RoomConnID, chan<- interface{}]

var roomMembers *xsync.MapOf[string, RoomMembers] = xsync.NewMapOf[string, RoomMembers]()

type UserConns = *xsync.MapOf[chan<- interface{}, UserConnInfo]

var userConns *xsync.MapOf[uuid.UUID, UserConns] = xsync.NewMapOf[uuid.UUID, UserConns]()

func RegisterConnection(
	roomId string, connId RoomConnID, userToken string, writeChannel chan<- interface{},
) (members RoomMembers, previousConnectionExisted bool) {
	members, _ = roomMembers.LoadOrStore(roomId, xsync.NewMapOf[RoomConnID, chan<- interface{}]())
	oldWriteChannel, previousConnectionExisted := members.LoadAndStore(connId, writeChannel)
	if previousConnectionExisted {
		oldWriteChannel <- WsInternalClientReconnect
	}
	connections, _ := userConns.LoadOrStore(connId.UserID, xsync.NewMapOf[chan<- interface{}, UserConnInfo]())
	connections.Store(writeChannel, UserConnInfo{RoomID: roomId, Token: userToken})
	if os.Getenv("CONCINNITY_DEBUG_CONNECTIONS") == "true" {
		log.Printf("C: Client ID: %s | Room %s members: %v\n", connId.ClientID, roomId, members.Size())
		log.Printf("C: Client ID: %s | User connections: %v\n", connId.ClientID, connections.Size())
	}
	return members, previousConnectionExisted
}

func UnregisterConnection(
	roomId string, connId RoomConnID, members RoomMembers, writeChannel chan<- interface{},
) {
	userConns.Compute(connId.UserID, func(value UserConns, loaded bool) (UserConns, bool) {
		value.Delete(writeChannel)
		if os.Getenv("CONCINNITY_DEBUG_CONNECTIONS") == "true" {
			log.Printf("DC: Client ID: %s | User connections: %v\n", connId.ClientID, value.Size())
		}
		return value, value.Size() == 0 // Delete user if no connections left
	})
	members.Compute(connId, func(value chan<- interface{}, loaded bool) (chan<- interface{}, bool) {
		if os.Getenv("CONCINNITY_DEBUG_CONNECTIONS") == "true" {
			size := members.Size()
			if value == writeChannel {
				size--
			}
			log.Printf("DC: Client ID: %s | Room %s members: %v\n", connId.ClientID, roomId, size)
		}
		return value, value == writeChannel // Delete only if this is the current i.e. right connection
	})
}

func PurgeExpiredDataTask() {
	for {
		time.Sleep(10 * time.Minute)
		if _, err := purgeExpiredPasswordResetTokensStmt.Exec(); err != nil {
			log.Println("Failed to purge expired password reset tokens!", err)
		}
		CleanInactiveRooms()
	}
}

func CleanInactiveRooms() {
	rows, err := findInactiveRoomsStmt.Query()
	if err != nil {
		log.Println("Failed to find inactive rooms!", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			log.Println("Failed to scan inactive room!", err)
			continue
		}

		if members, ok := roomMembers.Load(id); !ok || members.Size() == 0 {
			result, err := deleteRoomStmt.Exec(id)
			if err != nil {
				log.Println("Failed to delete inactive room!", err)
			} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
				log.Println("Failed to delete inactive room!", err)
			} else {
				roomMembers.Delete(id)
			}
		}
	}
	if rows.Err() != nil {
		log.Println("Failed to scan inactive room!", err)
	}
}
