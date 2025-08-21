package main

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func CreateSqlTables() {
	if _, err := db.Exec(translate(`BEGIN;

CREATE TABLE IF NOT EXISTS avatars (
  hash VARCHAR(64) NOT NULL PRIMARY KEY,
	data LONGBLOB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());

CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(16) NOT NULL UNIQUE,
	password VARCHAR(100) NOT NULL,
	email VARCHAR(319) NOT NULL UNIQUE,
	id UUID NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	verified BOOLEAN NOT NULL DEFAULT FALSE,
	avatar VARCHAR(64) NULL DEFAULT NULL REFERENCES avatars(hash) ON DELETE RESTRICT);

CREATE TABLE IF NOT EXISTS tokens (
	token VARCHAR(128) NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE);

CREATE TABLE IF NOT EXISTS rooms (
	id VARCHAR(24) NOT NULL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	type VARCHAR(24) NOT NULL, /* local_file, remote_file */
	target VARCHAR(1024) NOT NULL, /* carries information like file name, YouTube ID, etc */
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
	data MEDIUMTEXT NOT NULL,
	PRIMARY KEY (room_id, name));
CREATE INDEX IF NOT EXISTS subtitles_room_id_idx ON subtitles (room_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());

COMMIT;`)); err != nil {
		log.Fatalln("Failed to create tables and indexes!", err)
	}
}

func UpgradeSqlTables() {
	log.Println("Upgrading database schema...")
	tokenFKeyName := "tokens_user_id_fkey"
	if config.Database == "mysql" {
		tokenFKeyName = "tokens_ibfk_1" // MySQL uses different names for foreign keys
	}
	avatarFKeyName := "users_avatar_fkey"
	if config.Database == "mysql" {
		avatarFKeyName = "users_ibfk_1"
	}
	if _, err := db.Exec(translate(`BEGIN;

-- Upgrading from concinnity 1.0.0
ALTER TABLE tokens DROP CONSTRAINT ` + tokenFKeyName + `;
ALTER TABLE tokens ADD CONSTRAINT ` + tokenFKeyName + ` FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Upgrading from concinnity 1.0.1
-- [#Postgres] ALTER TABLE users DROP CONSTRAINT IF EXISTS ` + avatarFKeyName + `;
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS avatar VARCHAR(64) NULL DEFAULT NULL,
-- [#Postgres]	ADD CONSTRAINT ` + avatarFKeyName + ` FOREIGN KEY (avatar) REFERENCES avatars(hash) ON DELETE RESTRICT;
-- [#MySQL]	    ADD CONSTRAINT FOREIGN KEY IF NOT EXISTS ` + avatarFKeyName + ` (avatar) REFERENCES avatars(hash) ON DELETE RESTRICT;
-- [#Postgres] ALTER TABLE rooms ALTER COLUMN target TYPE VARCHAR(1024);
-- [#MySQL]    ALTER TABLE rooms MODIFY COLUMN target VARCHAR(1024) NOT NULL;

COMMIT;`)); err != nil {
		log.Fatalln("Failed to run database schema upgrade!", err)
	}
}

var (
	findUserByTokenStmt       *sql.Stmt
	findUserByNameOrEmailStmt *sql.Stmt
	findUserByUsernameStmt    *sql.Stmt
	findUserByEmailStmt       *sql.Stmt
	findUserProfilesByIdStmt  *sql.Stmt
	createUserStmt            *sql.Stmt
	updateUserPasswordStmt    *sql.Stmt
	updateUserUsernameStmt    *sql.Stmt
	updateUserEmailStmt       *sql.Stmt
	updateUserAvatarStmt      *sql.Stmt
	deleteUserStmt            *sql.Stmt

	insertTokenStmt *sql.Stmt
	deleteTokenStmt *sql.Stmt

	insertPasswordResetTokenStmt        *sql.Stmt
	findRecentPasswordResetTokensStmt   *sql.Stmt
	findUserByPasswordResetTokenStmt    *sql.Stmt
	deletePasswordResetTokenStmt        *sql.Stmt
	purgeExpiredPasswordResetTokensStmt *sql.Stmt

	findAvatarByHashStmt *sql.Stmt
	insertAvatarStmt     *sql.Stmt
	deleteAvatarStmt     *sql.Stmt

	insertRoomStmt         *sql.Stmt
	findRoomStmt           *sql.Stmt
	findRoomModifyTimeStmt *sql.Stmt // MySQL specific, complementing updateRoomStmt
	findInactiveRoomsStmt  *sql.Stmt
	updateRoomStmt         *sql.Stmt
	updateRoomModifiedStmt *sql.Stmt // MySQL specific, complementing insertChatMessageStmt
	updateRoomStateStmt    *sql.Stmt
	deleteRoomStmt         *sql.Stmt

	findChatMessagesByRoomStmt *sql.Stmt
	insertChatMessageStmt      *sql.Stmt

	findSubtitlesByRoomStmt *sql.Stmt
	findSubtitleStmt        *sql.Stmt
	insertSubtitleStmt      *sql.Stmt
	deleteRoomSubtitlesStmt *sql.Stmt // MySQL specific, complementing updateRoomStmt
)

func PrepareSqlStatements() {
	findUserByTokenStmt = prepareQuery("SELECT username, password, email, id, users.created_at " +
		"AS user_created_at, verified, avatar, token, tokens.created_at AS token_created_at FROM tokens " +
		"JOIN users ON tokens.user_id = users.id WHERE token = $1;")
	findUserByNameOrEmailStmt = prepareQuery("SELECT username, password, email, id, created_at, verified, avatar FROM users " +
		"WHERE username = $1 OR email = $2 LIMIT 1;")
	findUserByUsernameStmt = prepareQuery("SELECT username, password, email, id, created_at, verified, avatar FROM users " +
		"WHERE username = $1 LIMIT 1;")
	findUserByEmailStmt = prepareQuery("SELECT username, password, email, id, created_at, verified, avatar FROM users " +
		"WHERE email = $1 LIMIT 1;")
	if config.Database != "mysql" {
		findUserProfilesByIdStmt = prepareQuery("SELECT id, username, avatar FROM users WHERE id = ANY($1);")
	}
	createUserStmt = prepareQuery("INSERT INTO users (username, password, email, id, verified) VALUES ($1, $2, $3, $4, $5);")
	updateUserPasswordStmt = prepareQuery("UPDATE users SET password = $1 WHERE id = $2;")
	updateUserUsernameStmt = prepareQuery("UPDATE users SET username = $1 WHERE id = $2;")
	updateUserEmailStmt = prepareQuery("UPDATE users SET email = $1 WHERE id = $2;")
	updateUserAvatarStmt = prepareQuery("UPDATE users SET avatar = $1 WHERE id = $2;")
	deleteUserStmt = prepareQuery("DELETE FROM users WHERE id = $1;")

	insertTokenStmt = prepareQuery("INSERT INTO tokens (token, created_at, user_id) VALUES ($1, $2, $3);")
	deleteTokenStmt = prepareQuery("DELETE FROM tokens WHERE token = $1 RETURNING user_id;")

	insertPasswordResetTokenStmt = prepareQuery(
		"INSERT INTO password_reset_tokens (user_id) VALUES ($1) RETURNING id, user_id, created_at;")
	findRecentPasswordResetTokensStmt = prepareQuery(
		`SELECT id, user_id, created_at FROM password_reset_tokens
		WHERE created_at > NOW() - INTERVAL '2 minutes' AND user_id = $1;`)
	findUserByPasswordResetTokenStmt = prepareQuery(
		`SELECT users.id, users.username, password_reset_tokens.created_at
		FROM password_reset_tokens JOIN users ON password_reset_tokens.user_id = users.id
		WHERE password_reset_tokens.id = $1;`)
	deletePasswordResetTokenStmt = prepareQuery(
		"DELETE FROM password_reset_tokens WHERE id = $1 RETURNING user_id, created_at;")
	purgeExpiredPasswordResetTokensStmt = prepareQuery(
		"DELETE FROM password_reset_tokens WHERE created_at < NOW() - INTERVAL '10 minutes';")

	findAvatarByHashStmt = prepareQuery("SELECT hash, data, created_at FROM avatars WHERE hash = $1;")
	if config.Database == "mysql" {
		insertAvatarStmt = prepareQuery("INSERT IGNORE INTO avatars (hash, data) VALUES (?, ?);")
	} else {
		insertAvatarStmt = prepareQuery("INSERT INTO avatars (hash, data) VALUES ($1, $2) ON CONFLICT (hash) DO NOTHING;")
	}
	deleteAvatarStmt = prepareQuery("DELETE FROM avatars WHERE hash = $1;")

	insertRoomStmt = prepareQuery("INSERT INTO rooms (id, type, target) " +
		"VALUES ($1, $2, $3);")
	findRoomStmt = prepareQuery("SELECT id, created_at, modified_at, type, target, " +
		"paused, speed, timestamp, last_action FROM rooms WHERE id = $1;")
	findInactiveRoomsStmt = prepareQuery("SELECT id FROM rooms WHERE modified_at < NOW() - INTERVAL '10 minutes';")
	if config.Database == "mysql" {
		deleteRoomSubtitlesStmt = prepareQuery("DELETE FROM subtitles WHERE room_id = ?;")
		updateRoomStmt = prepareQuery(`UPDATE rooms
  		SET type = $2, target = $3, modified_at = NOW(),
					paused = true, speed = 1, timestamp = 0, last_action = NOW()
			WHERE id = $1;`)
		findRoomModifyTimeStmt = prepareQuery("SELECT created_at, modified_at FROM rooms WHERE id = $1;")
	} else {
		updateRoomStmt = prepareQuery(`
			WITH subs AS (
				DELETE FROM subtitles WHERE room_id = $1
			) UPDATE rooms
  			SET type = $2, target = $3, modified_at = NOW(),
						paused = true, speed = 1, timestamp = 0, last_action = NOW()
				WHERE id = $1
				RETURNING created_at, modified_at;`)
	}
	updateRoomStateStmt = prepareQuery("UPDATE rooms SET " +
		"paused = $2, speed = $3, timestamp = $4, last_action = $5, modified_at = NOW() WHERE id = $1;")
	deleteRoomStmt = prepareQuery("DELETE FROM rooms WHERE id = $1;")

	findChatMessagesByRoomStmt = prepareQuery("SELECT id, user_id, timestamp, message FROM chats WHERE room_id = $1;")
	if config.Database == "mysql" {
		updateRoomModifiedStmt = prepareQuery("UPDATE rooms SET modified_at = NOW() WHERE id = ?;")
		insertChatMessageStmt = prepareQuery(
			"INSERT INTO chats (room_id, user_id, message) VALUES (?, ?, ?) RETURNING id, timestamp;")
	} else {
		insertChatMessageStmt = prepareQuery(`
			WITH rooms AS (
  			UPDATE rooms SET modified_at = NOW() WHERE id = $1
			) INSERT INTO chats (room_id, user_id, message) VALUES ($1, $2, $3) RETURNING id, timestamp;`)
	}

	findSubtitlesByRoomStmt = prepareQuery("SELECT name FROM subtitles WHERE room_id = $1;")
	findSubtitleStmt = prepareQuery("SELECT data FROM subtitles WHERE room_id = $1 AND name = $2;")
	insertSubtitleStmt = prepareQuery(`
		INSERT INTO subtitles (room_id, name, data) VALUES ($1, $2, $3)
  	ON CONFLICT (room_id, name) DO UPDATE SET data = $3;
	`)
}

func translate(query string) string {
	if config.Database == "mysql" {
		query = regexp.MustCompile(`\$\d+`).ReplaceAllString(query, "?")
		query = strings.ReplaceAll(query, "TIMESTAMPTZ", "TIMESTAMP")
		query = strings.ReplaceAll(query, "GENERATED ALWAYS AS IDENTITY", "AUTO_INCREMENT")
		query = regexp.MustCompile(`INTERVAL '(\d+) (\w+)s'`).ReplaceAllString(query, "INTERVAL $1 $2")
		query = regexp.MustCompile(`ON CONFLICT \([^)]+\) DO UPDATE SET`).ReplaceAllString(query, "ON DUPLICATE KEY UPDATE")
		// DELETE ... RETURNING works only with MariaDB 10.0+ (not MySQL!)
		query = strings.ReplaceAll(query, "gen_random_uuid", "UUID") // UUID_v7() is MariaDB 11.7+ only!
		// ADD COLUMN IF NOT EXISTS works only with MariaDB 10.0+ (not MySQL!)
		// ADD CONSTRAINT IF NOT EXISTS works only with MariaDB 10.0+ (not MySQL!)
		query = strings.ReplaceAll(query, "-- [#MySQL]", "")
	} else {
		// ADD CONSTRAINT IF NOT EXISTS is unsupported by PostgreSQL.
		query = strings.ReplaceAll(query, "LONGBLOB", "BYTEA")
		query = strings.ReplaceAll(query, "MEDIUMTEXT", "TEXT")
		query = strings.ReplaceAll(query, "-- [#Postgres]", "")
	}
	return query
}

func prepareQuery(query string) *sql.Stmt {
	query = translate(query)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatalln("failed to build SQL query:", query, err)
	}
	return stmt
}

func UpdateRoom(id string, roomType string, target string) (createdAt, modifiedAt time.Time, err error) {
	if config.Database == "mysql" {
		tx, err := db.Begin()
		if err != nil {
			return createdAt, modifiedAt, err
		}
		defer tx.Rollback()
		_, err = tx.Stmt(deleteRoomSubtitlesStmt).Exec(id)
		if err != nil {
			return createdAt, modifiedAt, err
		}
		result, err := tx.Stmt(updateRoomStmt).Exec(roomType, target, id)
		if err != nil {
			return createdAt, modifiedAt, err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return createdAt, modifiedAt, err
		} else if rowsAffected == 0 {
			return createdAt, modifiedAt, sql.ErrNoRows
		}
		err = tx.Stmt(findRoomModifyTimeStmt).QueryRow(id).Scan(&createdAt, &modifiedAt)
		if err != nil {
			return createdAt, modifiedAt, err
		}
		if err = tx.Commit(); err != nil {
			return createdAt, modifiedAt, err
		}
		return createdAt, modifiedAt, err
	}
	err = updateRoomStmt.QueryRow(id, roomType, target).Scan(&createdAt, &modifiedAt)
	return
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

func InsertChatMessage(roomId string, userId *uuid.UUID, message string) (id int, timestamp time.Time, err error) {
	if config.Database == "mysql" {
		tx, err := db.Begin()
		if err != nil {
			return id, timestamp, err
		}
		defer tx.Rollback()
		_, err = tx.Stmt(updateRoomModifiedStmt).Exec(roomId)
		if err != nil {
			return id, timestamp, err
		}
		err = tx.Stmt(insertChatMessageStmt).QueryRow(roomId, userId, message).Scan(&id, &timestamp)
		if err != nil {
			return id, timestamp, err
		}
		if err = tx.Commit(); err != nil {
			return id, timestamp, err
		}
		return id, timestamp, err
	}
	err = insertChatMessageStmt.QueryRow(roomId, userId, message).Scan(&id, &timestamp)
	return
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
