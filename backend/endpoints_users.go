package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"io"
	"net/http"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
	"github.com/go-sql-driver/mysql"
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
	avatarData := bytes.NewBuffer(avatar.Data)
	if r.URL.Query().Get("size") == "256" {
		// Decode original image
		originalImage, err := nativewebp.Decode(avatarData)
		if err != nil {
			http.Error(w, errorJson("Failed to decode avatar image!"), http.StatusUnprocessableEntity)
			return
		}
		// Resize image
		resizedImage := originalImage
		if originalImage.Bounds().Dx() > 256 || originalImage.Bounds().Dy() > 256 {
			resizedImage = imaging.Resize(originalImage, 256, 256, imaging.Lanczos)
			// Encode image
			avatarData.Reset()
			err = nativewebp.Encode(avatarData, resizedImage, &nativewebp.Options{
				UseExtendedFormat: true,
			})
			if err != nil {
				handleInternalServerError(w, err)
				return
			}
		}
	}
	// Return the avatar
	w.Header().Set("Content-Type", "image/webp")
	http.ServeContent(w, r, avatar.Hash+".webp", avatar.CreatedAt, bytes.NewReader(avatarData.Bytes()))
}

func ChangeAvatarEndpoint(w http.ResponseWriter, r *http.Request) {
	user, token := IsAuthenticatedHTTP(w, r)
	if token == nil {
		return
	}
	// Read the body
	var data []byte = nil
	var hash string = ""
	if r.Body != nil {
		avatarData := new(bytes.Buffer)
		if n, err := io.CopyN(avatarData, r.Body, 1024*1024*16); err != nil && !errors.Is(err, io.EOF) {
			handleInternalServerError(w, err)
			return
		} else if n == 1024*1024*16 { // 16 MB limit
			http.Error(w, errorJson("Avatar data too large! Maximum size is 16 MB."), http.StatusBadRequest)
			return
		} else if n > 0 {
			// Decode the image
			originalImage, _, err := image.Decode(avatarData)
			if err != nil {
				http.Error(w, errorJson("Failed to decode avatar image! Supported formats: WebP, PNG, JPEG"),
					http.StatusBadRequest)
				return
			}
			// Crop the image
			croppedImage := originalImage
			if originalImage.Bounds().Dx() != originalImage.Bounds().Dy() {
				res := min(originalImage.Bounds().Dx(), originalImage.Bounds().Dy(), 4096)
				croppedImage = imaging.Fill(originalImage, res, res, imaging.Center, imaging.Lanczos)
			} else if originalImage.Bounds().Dx() > 4096 || originalImage.Bounds().Dy() > 4096 {
				croppedImage = imaging.Resize(originalImage, 4096, 4096, imaging.Lanczos)
			}
			// Encode the image
			avatarData.Reset()
			err = nativewebp.Encode(avatarData, croppedImage, &nativewebp.Options{
				UseExtendedFormat: true,
			})
			if err != nil {
				handleInternalServerError(w, err)
				return
			}
			data = avatarData.Bytes()
			hashBytes := sha256.Sum256(data)
			hash = hex.EncodeToString(hashBytes[:])
		}
	}

	tx, err := db.Begin()
	if err != nil {
		handleInternalServerError(w, err)
		return
	}
	defer tx.Rollback()

	// Insert new avatar (ignore if avatar already exists)
	if hash != "" {
		if _, err := tx.Stmt(insertAvatarStmt).Exec(hash, data); err != nil {
			handleInternalServerError(w, err)
			return
		}
	}
	// Update user avatar
	avatarHash := sql.NullString{Valid: hash != "", String: hash}
	if result, err := tx.Stmt(updateUserAvatarStmt).Exec(avatarHash, token.UserID); err != nil {
		handleInternalServerError(w, err)
		return
	} else if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		handleInternalServerError(w, err) // nil err solved by Ostrich algorithm
		return
	}
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		handleInternalServerError(w, err)
		return
	}
	// Delete old avatar
	if user.Avatar != nil {
		_, err := deleteAvatarStmt.Exec(*user.Avatar)
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
			// Do nothing
		} else if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1451 {
			// Do nothing
		} else if err != nil {
			handleInternalServerError(w, err)
			return
		}
	}

	w.Write([]byte("{\"success\":true}"))
}

func GetUserProfilesEndpoint(w http.ResponseWriter, r *http.Request) {
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
			"SELECT id, username, avatar FROM users WHERE id IN (" + placeholders + ");").Query(mysqlArr...)
	} else {
		rows, err = findUserProfilesByIdStmt.Query(pq.Array(requestedIds))
	}
	if err != nil {
		handleInternalServerError(w, err)
		return
	}

	defer rows.Close()
	usernames := make(map[string]UserProfile)
	for rows.Next() {
		var id uuid.UUID
		var userProfile UserProfile
		if err = rows.Scan(&id, &userProfile.Username, &userProfile.Avatar); err != nil {
			handleInternalServerError(w, err)
			return
		}
		usernames[id.String()] = userProfile
	}
	if rows.Err() != nil {
		handleInternalServerError(w, err)
		return
	}

	json.NewEncoder(w).Encode(usernames)
}
