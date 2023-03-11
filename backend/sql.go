package main

import (
	"database/sql"
	"log"
)

var findUserByTokenStmt *sql.Stmt
var findUserByNameOrEmailStmt *sql.Stmt
var findUserByUsernameStmt *sql.Stmt
var findUserByEmailStmt *sql.Stmt
var createUserStmt *sql.Stmt

var insertTokenStmt *sql.Stmt
var deleteTokenStmt *sql.Stmt

var insertRoomStmt *sql.Stmt
var findRoomByIdStmt *sql.Stmt

const findUserByTokenQuery = "SELECT username, password, email, tokens.id AS id, users.createdAt " +
	"AS userCreatedAt, verified, token, tokens.createdAt AS tokenCreatedAt FROM tokens " +
	"JOIN users ON tokens.id = users.id WHERE token = $1;"
const findUserByNameOrEmailQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE username = $1 OR email = $2 LIMIT 1;"
const findUserByUsernameQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE username = $1 LIMIT 1;"
const findUserByEmailQuery = "SELECT username, password, email, id, createdAt, verified FROM users " +
	"WHERE email = $1 LIMIT 1;"
const createUserQuery = "INSERT INTO users (username, password, email, id) VALUES ($1, $2, $3, $4);"

const insertTokenQuery = "INSERT INTO tokens (token, createdAt, id) VALUES ($1, $2, $3);"
const deleteTokenQuery = "DELETE FROM tokens WHERE token = $1;"

const insertRoomQuery = "INSERT INTO rooms (id, type, title, extra, members) " +
	"VALUES ($1, $2, $3, $4, $5);"
const findRoomByIdQuery = "SELECT * FROM rooms WHERE id = $1;"

// TODO: Rename token.id to userId?
// TODO: UUIDs are yucky, can we use nanoid instead for rooms?
// TODO: Do we need user IDs even? Isn't username sufficient? Should usernames even be unique?
const createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(16) UNIQUE,
	password VARCHAR(100),
	email TEXT UNIQUE,
	id UUID PRIMARY KEY,
	createdAt TIMESTAMPTZ DEFAULT NOW(),
	verified BOOLEAN DEFAULT FALSE);`
const createTokensTableQuery = `CREATE TABLE IF NOT EXISTS tokens (
	token VARCHAR(128) PRIMARY KEY,
	createdAt TIMESTAMPTZ DEFAULT NOW(),
	id UUID);`
const createRoomsTableQuery = `CREATE TABLE IF NOT EXISTS rooms (
	id UUID PRIMARY KEY,
	type VARCHAR(24),
	title VARCHAR(200),
	extra VARCHAR(200),
	chat VARCHAR(2100)[] DEFAULT '{}',
	members VARCHAR(16)[] DEFAULT '{}',
	paused BOOLEAN DEFAULT TRUE,
	timestamp INTEGER DEFAULT 0,
	lastActionTime TIMESTAMPTZ DEFAULT NOW(),
	createdAt TIMESTAMPTZ DEFAULT NOW());`

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
	findRoomByIdStmt, err = db.Prepare(findRoomByIdQuery)
	if err != nil {
		log.Panicln("Failed to prepare query to find room by id!", err)
	}
}
