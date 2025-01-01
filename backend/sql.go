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
var insertChatMessageRoomStmt *sql.Stmt
var updateRoomStateStmt *sql.Stmt
var deleteRoomStmt *sql.Stmt

const findUserByTokenQuery = "SELECT username, password, email, id, users.createdAt " +
	"AS userCreatedAt, verified, token, tokens.createdAt AS tokenCreatedAt FROM tokens " +
	"JOIN users ON tokens.userId = users.id WHERE token = $1;"
const findUserByNameOrEmailQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE username = $1 OR email = $2 LIMIT 1;"
const findUserByUsernameQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE username = $1 LIMIT 1;"
const findUserByEmailQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE email = $1 LIMIT 1;"
const findUsernamesByIdQuery = "SELECT id, username FROM users WHERE id = ANY($1);"
const createUserQuery = "INSERT INTO users (username, password, email, id) VALUES ($1, $2, $3, $4);"

const insertTokenQuery = "INSERT INTO tokens (token, createdAt, userId) VALUES ($1, $2, $3);"
const deleteTokenQuery = "DELETE FROM tokens WHERE token = $1;"

const insertRoomQuery = "INSERT INTO rooms (id, type, target) " +
	"VALUES ($1, $2, $3);"
const findRoomQuery = "SELECT id, createdAt, modifiedAt, type, target, chat, " +
	"paused, speed, timestamp, lastAction FROM rooms WHERE id = $1;"
const findInactiveRoomsQuery = "SELECT id FROM rooms WHERE modifiedAt < NOW() - INTERVAL '10 minutes';"
const updateRoomQuery = "UPDATE rooms SET type = $2, target = $3, modifiedAt = NOW(), " +
	"paused = true, speed = 1, timestamp = 0, lastAction = NOW() WHERE id = $1 " +
	"RETURNING createdAt, modifiedAt;"
const insertChatMessageRoomQuery = "UPDATE rooms SET chat = chat || $2::jsonb, modifiedAt = NOW() WHERE id = $1;"
const updateRoomStateQuery = "UPDATE rooms SET " +
	"paused = $2, speed = $3, timestamp = $4, lastAction = $5, modifiedAt = NOW() WHERE id = $1;"
const deleteRoomQuery = "DELETE FROM rooms WHERE id = $1;"

const createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(16) UNIQUE,
	password VARCHAR(100),
	email VARCHAR(319) UNIQUE,
	id UUID PRIMARY KEY,
	createdAt TIMESTAMPTZ DEFAULT NOW(),
	verified BOOLEAN DEFAULT FALSE);`
const createTokensTableQuery = `CREATE TABLE IF NOT EXISTS tokens (
	token VARCHAR(128) PRIMARY KEY,
	createdAt TIMESTAMPTZ DEFAULT NOW(),
	userId UUID REFERENCES users(id));`
const createRoomsTableQuery = `CREATE TABLE IF NOT EXISTS rooms (
	id VARCHAR(24) PRIMARY KEY,
	createdAt TIMESTAMPTZ DEFAULT NOW(),
	modifiedAt TIMESTAMPTZ DEFAULT NOW(),
	type VARCHAR(24), /* local_file, remote_file */
	target VARCHAR(200), /* carries information like file name, YouTube ID, etc */
	chat JSONB[] DEFAULT '{}',
	paused BOOLEAN DEFAULT TRUE,
	speed INTEGER DEFAULT 1,
	timestamp INTEGER DEFAULT 0,
	lastAction TIMESTAMPTZ DEFAULT NOW());`

func CreateSqlTables() {
	_, err := db.Exec(createUsersTableQuery)
	if err != nil {
		log.Panicln("Failed to create users table!", err)
	}
	_, err = db.Exec(createTokensTableQuery)
	if err != nil {
		log.Panicln("Failed to create tokens table!", err)
	}
	_, err = db.Exec(createRoomsTableQuery)
	if err != nil {
		log.Panicln("Failed to create rooms table!", err)
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
	insertChatMessageRoomStmt, err = db.Prepare(insertChatMessageRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to insert chat message in room!", err)
	}
	updateRoomStateStmt, err = db.Prepare(updateRoomStateQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to update room state!", err)
	}
	deleteRoomStmt, err = db.Prepare(deleteRoomQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to delete room!", err)
	}
}
