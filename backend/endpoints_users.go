package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func GetAvatarEndpoint(w http.ResponseWriter, r *http.Request) {
	// This endpoint does not require authentication
	if len(r.PathValue("hash")) != 64 {
		http.Error(w, errorJson("Invalid avatar hash!"), http.StatusBadRequest)
		return
	} else if r.URL.Query().Get("size") != "" &&
		r.URL.Query().Get("size") != "256" &&
		r.URL.Query().Get("size") != "4096" {
		http.Error(w, errorJson("Invalid size parameter! Supported sizes: 256, 4096"),
			http.StatusBadRequest)
		return
	}
	// Retrieve avatar from the database
	var avatar Avatar
	err := findAvatarByHashStmt.QueryRow(r.PathValue("hash")).Scan(
		&avatar.Hash, &avatar.Data, &avatar.UpdatedAt, &avatar.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Avatar not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// TODO: If ?size=256, downscale the avatar
	// Return the avatar
	w.Header().Set("Content-Type", "image/webp")
	http.ServeContent(w, r, avatar.Hash+".webp", avatar.UpdatedAt, bytes.NewReader(avatar.Data))
}

func GetUsernamesEndpoint(w http.ResponseWriter, r *http.Request) {
	_, token := IsAuthenticatedHTTP(w, r)
	if token == nil {
		return
	}
	requestedIds := r.URL.Query()["id"]
	if len(requestedIds) == 0 {
		http.Error(w, errorJson("No user IDs provided!"), http.StatusBadRequest)
		return
	}
	ids := make([]uuid.UUID, len(requestedIds))
	var err error
	for i, id := range requestedIds {
		ids[i], err = uuid.Parse(id)
		if err != nil {
			http.Error(w, errorJson("Invalid user ID(s) provided!"), http.StatusBadRequest)
			return
		}
	}

	var rows *sql.Rows
	if config.Database == "mysql" {
		placeholders := strings.Repeat("?,", len(requestedIds))
		placeholders = placeholders[:len(placeholders)-1]
		mysqlArr := make([]interface{}, len(requestedIds))
		for i, id := range requestedIds {
			mysqlArr[i] = id
		}
		rows, err = prepareQuery(
			"SELECT id, username FROM users WHERE id IN (" + placeholders + ");").Query(mysqlArr...)
	} else {
		rows, err = findUsernamesByIdStmt.Query(pq.Array(requestedIds))
	}
	if err != nil {
		handleInternalServerError(w, err)
		return
	}

	defer rows.Close()
	usernames := make(map[string]string)
	for rows.Next() {
		var id uuid.UUID
		var username string
		if err = rows.Scan(&id, &username); err != nil {
			handleInternalServerError(w, err)
			return
		}
		usernames[id.String()] = username
	}
	if rows.Err() != nil {
		handleInternalServerError(w, err)
		return
	}

	json.NewEncoder(w).Encode(usernames)
}
