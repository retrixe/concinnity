package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

func errorJson(err string) string {
	json, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: err})
	return string(json)
}

func handleInternalServerError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
}

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
		var id []byte
		err = rows.Scan(&id)
		if err != nil {
			log.Println("Failed to scan inactive room!", err)
			continue
		}
		// TODO: Check if there are no members in the room.
		_, err = deleteRoomStmt.Exec(id)
		if err != nil {
			log.Println("Failed to delete inactive room!", err)
			continue
		}
	}
}

// GenerateSalt returns a 16-character salt readable in UTF-8 format as well.
func GenerateSalt() []byte {
	saltBytes := make([]byte, 12)
	_, _ = rand.Read(saltBytes)
	salt := base64.RawStdEncoding.EncodeToString(saltBytes)
	return []byte(salt)
}

func HashPassword(password string, salt []byte) string {
	params := "$argon2id$v=19$m=51200,t=1,p=4$" // Currently fixed only.
	key := argon2.IDKey([]byte(password), salt, 1, 51200, 4, 32)
	return params + base64.RawStdEncoding.EncodeToString(salt) +
		"$" + base64.RawStdEncoding.EncodeToString(key)
}

func ComparePassword(password string, hash string) bool {
	encodeSplit := strings.Split(hash, "$")
	salt, _ := base64.RawStdEncoding.DecodeString(encodeSplit[len(encodeSplit)-2])
	key := argon2.IDKey([]byte(password), salt, 1, 51200, 4, 32)
	hashValue := encodeSplit[len(encodeSplit)-1]
	return hashValue == base64.RawStdEncoding.EncodeToString(key)
}
