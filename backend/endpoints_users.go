package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
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
		&avatar.Hash, &avatar.Data, &avatar.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, errorJson("Avatar not found!"), http.StatusNotFound)
		return
	} else if err != nil {
		handleInternalServerError(w, err)
		return
	}
	// If ?size=256, downscale the avatar
	avatarData := avatar.Data
	if r.URL.Query().Get("size") == "256" {
		// Decode original image
		originalImage, err := nativewebp.Decode(bytes.NewReader(avatarData))
		if err != nil {
			http.Error(w, errorJson("Failed to decode avatar image!"), http.StatusUnprocessableEntity)
			return
		}
		// Resize image
		resizedImage := originalImage
		if originalImage.Bounds().Dx() > 256 || originalImage.Bounds().Dy() > 256 {
			resizedImage = imaging.Resize(originalImage, 256, 256, imaging.Lanczos)
		}
		// Encode image
		newDataWriter := new(bytes.Buffer)
		err = nativewebp.Encode(newDataWriter, resizedImage, &nativewebp.Options{
			UseExtendedFormat: true,
		})
		if err != nil {
			http.Error(w, errorJson("Failed to encode avatar image!"), http.StatusInternalServerError)
			return
		}
		avatarData = newDataWriter.Bytes()
	}
	// Return the avatar
	w.Header().Set("Content-Type", "image/webp")
	http.ServeContent(w, r, avatar.Hash+".webp", avatar.CreatedAt, bytes.NewReader(avatarData))
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
