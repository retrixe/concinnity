package main

import (
	"database/sql"
	"log"
)

func CreateSqlTables() {
	if _, err := db.Exec(`BEGIN;

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

CREATE TABLE IF NOT EXISTS subtitles (
  room_id VARCHAR(24) NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  name VARCHAR(200) NOT NULL,
	data TEXT NOT NULL,
	PRIMARY KEY (room_id, name));
CREATE INDEX IF NOT EXISTS subtitles_room_id_idx ON subtitles (room_id);

COMMIT;`); err != nil {
		log.Panicln("Failed to create tables and indexes!", err)
	}
}

var (
	findUserByTokenStmt       *sql.Stmt
	findUserByNameOrEmailStmt *sql.Stmt
	findUserByUsernameStmt    *sql.Stmt
	findUserByEmailStmt       *sql.Stmt
	findUsernamesByIdStmt     *sql.Stmt
	createUserStmt            *sql.Stmt

	insertTokenStmt *sql.Stmt
	deleteTokenStmt *sql.Stmt

	insertRoomStmt        *sql.Stmt
	findRoomStmt          *sql.Stmt
	findInactiveRoomsStmt *sql.Stmt
	updateRoomStmt        *sql.Stmt
	updateRoomStateStmt   *sql.Stmt
	deleteRoomStmt        *sql.Stmt

	findChatMessagesByRoomStmt *sql.Stmt
	insertChatMessageStmt      *sql.Stmt

	findSubtitlesByRoomStmt *sql.Stmt
	findSubtitleStmt        *sql.Stmt
	insertSubtitleStmt      *sql.Stmt
)

func PrepareSqlStatements() {
	findUserByTokenStmt = prepareQuery("SELECT username, password, email, id, users.created_at " +
		"AS user_created_at, verified, token, tokens.created_at AS token_created_at FROM tokens " +
		"JOIN users ON tokens.user_id = users.id WHERE token = $1;")
	findUserByNameOrEmailStmt = prepareQuery("SELECT username, password, email, id, created_at, verified FROM users " +
		"WHERE username = $1 OR email = $2 LIMIT 1;")
	findUserByUsernameStmt = prepareQuery("SELECT username, password, email, id, created_at, verified FROM users " +
		"WHERE username = $1 LIMIT 1;")
	findUserByEmailStmt = prepareQuery("SELECT username, password, email, id, created_at, verified FROM users " +
		"WHERE email = $1 LIMIT 1;")
	findUsernamesByIdStmt = prepareQuery("SELECT id, username FROM users WHERE id = ANY($1);")
	createUserStmt = prepareQuery("INSERT INTO users (username, password, email, id, verified) VALUES ($1, $2, $3, $4, $5);")

	insertTokenStmt = prepareQuery("INSERT INTO tokens (token, created_at, user_id) VALUES ($1, $2, $3);")
	deleteTokenStmt = prepareQuery("DELETE FROM tokens WHERE token = $1;")

	insertRoomStmt = prepareQuery("INSERT INTO rooms (id, type, target) " +
		"VALUES ($1, $2, $3);")
	findRoomStmt = prepareQuery("SELECT id, created_at, modified_at, type, target, " +
		"paused, speed, timestamp, last_action FROM rooms WHERE id = $1;")
	findInactiveRoomsStmt = prepareQuery("SELECT id FROM rooms WHERE modified_at < NOW() - INTERVAL '10 minutes';")
	updateRoomStmt = prepareQuery(`
		WITH subs AS (
			DELETE FROM subtitles WHERE room_id = $1
		) UPDATE rooms
  		SET type = $2, target = $3, modified_at = NOW(),
					paused = true, speed = 1, timestamp = 0, last_action = NOW()
			WHERE id = $1
			RETURNING created_at, modified_at;`)
	updateRoomStateStmt = prepareQuery("UPDATE rooms SET " +
		"paused = $2, speed = $3, timestamp = $4, last_action = $5, modified_at = NOW() WHERE id = $1;")
	deleteRoomStmt = prepareQuery("DELETE FROM rooms WHERE id = $1;")

	findChatMessagesByRoomStmt = prepareQuery("SELECT id, user_id, timestamp, message FROM chats WHERE room_id = $1;")
	insertChatMessageStmt = prepareQuery(`
		WITH rooms AS (
  		UPDATE rooms SET modified_at = NOW() WHERE id = $1
		) INSERT INTO chats (room_id, user_id, message) VALUES ($1, $2, $3) RETURNING id, timestamp;`)

	findSubtitlesByRoomStmt = prepareQuery("SELECT name FROM subtitles WHERE room_id = $1;")
	findSubtitleStmt = prepareQuery("SELECT data FROM subtitles WHERE room_id = $1 AND name = $2;")
	insertSubtitleStmt = prepareQuery(`
		INSERT INTO subtitles (room_id, name, data) VALUES ($1, $2, $3)
  	ON CONFLICT (room_id, name) DO UPDATE SET data = $3;
	`)
}

func prepareQuery(query string) *sql.Stmt {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Panicln("failed to build SQL query: ", query)
	}
	return stmt
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

func FindSubtitlesByRoom(roomId string) ([]string, error) {
	names := make([]string, 0)
	nameRows, err := findSubtitlesByRoomStmt.Query(roomId)
	if err != nil {
		return nil, err
	}
	defer nameRows.Close()
	for nameRows.Next() {
		var name string
		if err = nameRows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err = nameRows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
