package main

import (
	"database/sql"
	"log"
)

var findUserByTokenStmt *sql.Stmt
var findUserByNameOrEmailStmt *sql.Stmt
var findUserByUsernameStmt *sql.Stmt
var findUserByEmailStmt *sql.Stmt
var insertUserStmt *sql.Stmt

var insertTokenStmt *sql.Stmt
var deleteTokenStmt *sql.Stmt

const findUserByTokenQuery = "SELECT username, password, email, tokens.id as id, token, createdAt FROM tokens " +
	"JOIN users ON tokens.id = users.id WHERE token = $1;"
const findUserByNameOrEmailQuery = "SELECT username, password, email, id FROM users " +
	"WHERE username = $1 OR email = $2 LIMIT 1;"
const findUserByUsernameQuery = "SELECT username, password, email, id FROM users " +
	"WHERE username = $1 LIMIT 1;"
const findUserByEmailQuery = "SELECT username, password, email, id FROM users " +
	"WHERE email = $1 LIMIT 1;"
const insertUserQuery = "INSERT INTO users (username, password, email, id) VALUES ($1, $2, $3, $4);"

const insertTokenQuery = "INSERT INTO tokens (token, createdAt, id) VALUES ($1, $2, $3);"
const deleteTokenQuery = "DELETE FROM tokens WHERE token = $1;"

const createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(16) UNIQUE,
	password VARCHAR(100),
	email TEXT UNIQUE,
	id UUID UNIQUE);`
const createTokensTableQuery = `CREATE TABLE IF NOT EXISTS tokens (
	token VARCHAR(128) UNIQUE,
	createdAt TIMESTAMPTZ,
	id UUID);`

// TODO rooms - members, id, chat, timestamp, paused, lastActionTime, createdAt

func CreateSqlTables() {
	_, err := db.Exec(createUsersTableQuery)
	if err != nil {
		log.Panicln("Failed to create users table!", err)
	}
	_, err = db.Exec(createTokensTableQuery)
	if err != nil {
		log.Panicln("Failed to create tokens table!", err)
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
	insertUserStmt, err = db.Prepare(insertUserQuery)
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
}
