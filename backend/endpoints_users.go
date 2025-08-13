package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

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
