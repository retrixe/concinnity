package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
)

type RoomConnID struct {
	UserID   uuid.UUID
	ClientID string
}

type RoomMembers = *xsync.MapOf[RoomConnID, chan<- interface{}]

var roomMembers *xsync.MapOf[string, RoomMembers] = xsync.NewMapOf[string, RoomMembers]()

type UserConns = *xsync.MapOf[chan<- interface{}, string]

var userConns *xsync.MapOf[uuid.UUID, UserConns] = xsync.NewMapOf[uuid.UUID, UserConns]()

func CleanInactiveRoomsTask() {
	for {
		time.Sleep(10 * time.Minute)
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
