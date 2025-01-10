package main

import (
	"database/sql"
	"log"
)

var findUserByTokenStmt *sql.Stmt
var findUserByNameOrEmailStmt *sql.Stmt
var findUserByUsernameStmt *sql.Stmt
var findUserByEmailStmt *sql.Stmt
var findUsernamesByIdStmt *sql.Stmt
var createUserStmt *sql.Stmt

var insertTokenStmt *sql.Stmt
var deleteTokenStmt *sql.Stmt

var insertRoomStmt *sql.Stmt
var findRoomStmt *sql.Stmt
var findInactiveRoomsStmt *sql.Stmt
var updateRoomStmt *sql.Stmt
var updateRoomStateStmt *sql.Stmt
var deleteRoomStmt *sql.Stmt

var findChatMessagesByRoomStmt *sql.Stmt
var insertChatMessageStmt *sql.Stmt

const findUserByTokenQuery = "SELECT username, password, email, id, users.created_at " +
	"AS user_created_at, verified, token, tokens.created_at AS token_created_at FROM tokens " +
	"JOIN users ON tokens.user_id = users.id WHERE token = $1;"
const findUserByNameOrEmailQuery = "SELECT username, password, email, id, created_at, verified FROM users " +
	"WHERE username = $1 OR email = $2 LIMIT 1;"
const findUserByUsernameQuery = "SELECT username, password, email, id, created_at, verified FROM users " +
	"WHERE username = $1 LIMIT 1;"
const findUserByEmailQuery = "SELECT username, password, email, id, created_at, verified FROM users " +
	"WHERE email = $1 LIMIT 1;"
const findUsernamesByIdQuery = "SELECT id, username FROM users WHERE id = ANY($1);"
const createUserQuery = "INSERT INTO users (username, password, email, id) VALUES ($1, $2, $3, $4);"

const insertTokenQuery = "INSERT INTO tokens (token, created_at, user_id) VALUES ($1, $2, $3);"
const deleteTokenQuery = "DELETE FROM tokens WHERE token = $1;"

const insertRoomQuery = "INSERT INTO rooms (id, type, target) " +
	"VALUES ($1, $2, $3);"
const findRoomQuery = "SELECT id, created_at, modified_at, type, target, " +
	"paused, speed, timestamp, last_action FROM rooms WHERE id = $1;"
const findInactiveRoomsQuery = "SELECT id FROM rooms WHERE modified_at < NOW() - INTERVAL '10 minutes';"
const updateRoomQuery = "UPDATE rooms SET type = $2, target = $3, modified_at = NOW(), " +
	"paused = true, speed = 1, timestamp = 0, last_action = NOW() WHERE id = $1 " +
	"RETURNING created_at, modified_at;"
const updateRoomStateQuery = "UPDATE rooms SET " +
	"paused = $2, speed = $3, timestamp = $4, last_action = $5, modified_at = NOW() WHERE id = $1;"
const deleteRoomQuery = "DELETE FROM rooms WHERE id = $1;"

const findChatMessagesByRoomQuery = "SELECT id, user_id, timestamp, message FROM chats WHERE room_id = $1;"
const insertChatMessageQuery = `WITH rooms AS (
  UPDATE rooms SET modified_at = NOW() WHERE id = $1
) INSERT INTO chats (room_id, user_id, message) VALUES ($1, $2, $3) RETURNING id, timestamp;`

const initialiseDatabaseQuery = `BEGIN;

CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(16) NOT NULL UNIQUE,
	password VARCHAR(100) NOT NULL,
	email VARCHAR(319) NOT NULL UNIQUE,
	id UUID NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	verified BOOLEAN NOT NULL DEFAULT FALSE);

CREATE TABLE IF NOT EXISTS tokens (
	token VARCHAR(128) NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	user_id UUID NOT NULL REFERENCES users(id));

CREATE TABLE IF NOT EXISTS rooms (
	id VARCHAR(24) NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	type VARCHAR(24) NOT NULL, /* local_file, remote_file */
	target VARCHAR(200) NOT NULL, /* carries information like file name, YouTube ID, etc */
	paused BOOLEAN NOT NULL DEFAULT TRUE,
	speed DECIMAL NOT NULL DEFAULT 1,
	timestamp DECIMAL NOT NULL DEFAULT 0,
	last_action TIMESTAMPTZ NOT NULL DEFAULT NOW());

CREATE TABLE IF NOT EXISTS chats (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	room_id VARCHAR(24) NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
	user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
	message TEXT NOT NULL,
	timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE INDEX IF NOT EXISTS chats_room_id_idx ON chats (room_id);
/* CREATE INDEX IF NOT EXISTS chats_timestamp_idx ON chats (timestamp); ORDER BY can't be so slow pfft */

COMMIT;`

func CreateSqlTables() {
	_, err := db.Exec(initialiseDatabaseQuery)
	if err != nil {
		log.Panicln("Failed to create tables and indexes!", err)
	}
}

func PrepareSqlStatements() {
	var err error

	findUserByTokenStmt, err = db.Prepare(findUserByTokenQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find user by token!", err)
	}
	findUserByNameOrEmailStmt, err = db.Prepare(findUserByNameOrEmailQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find user by username or email!", err)
	}
	findUserByUsernameStmt, err = db.Prepare(findUserByUsernameQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find user by username!", err)
	}
	findUserByEmailStmt, err = db.Prepare(findUserByEmailQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find user by email!", err)
	}
	findUsernamesByIdStmt, err = db.Prepare(findUsernamesByIdQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find usernames by ID!", err)
	}
	createUserStmt, err = db.Prepare(createUserQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to insert user!", err)
	}

	insertTokenStmt, err = db.Prepare(insertTokenQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to insert token!", err)
	}
	deleteTokenStmt, err = db.Prepare(deleteTokenQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to delete token!", err)
	}

	insertRoomStmt, err = db.Prepare(insertRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to insert room!", err)
	}
	findRoomStmt, err = db.Prepare(findRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find room!", err)
	}
	findInactiveRoomsStmt, err = db.Prepare(findInactiveRoomsQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find inactive rooms!", err)
	}
	updateRoomStmt, err = db.Prepare(updateRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to update room!", err)
	}
	updateRoomStateStmt, err = db.Prepare(updateRoomStateQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to update room state!", err)
	}
	deleteRoomStmt, err = db.Prepare(deleteRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to delete room!", err)
	}

	findChatMessagesByRoomStmt, err = db.Prepare(findChatMessagesByRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find chat messages by room ID!", err)
	}
	insertChatMessageStmt, err = db.Prepare(insertChatMessageQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to insert chat message!", err)
	}
}

func FindChatMessagesByRoom(id string) ([]ChatMessage, error) {
	chat := make([]ChatMessage, 0)
	chatRows, err := findChatMessagesByRoomStmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer chatRows.Close()
	for chatRows.Next() {
		msg := ChatMessage{}
		if err = chatRows.Scan(&msg.ID, &msg.UserID, &msg.Timestamp, &msg.Message); err != nil {
			return nil, err
		}
		chat = append(chat, msg)
	}
	if err = chatRows.Err(); err != nil {
		return nil, err
	}
	return chat, nil
}
