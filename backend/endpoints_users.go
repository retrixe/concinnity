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
	"slices"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Anecdotal samples:
// - Quality wise, 80 can be visually lacking with some images, 85 is fine, 90 is almost perfect
// - File size wise, 85 sits midway between 80-90 for similar quality to 90
const AVIF_QUALITY = 85

var VALID_AVATAR_SIZES = []string{"", "256", "4096"}

func GetAvatarEndpoint(w http.ResponseWriter, r *http.Request) {
	// This endpoint does not require authentication
	if len(r.PathValue("hash")) != 64 {
		http.Error(w, errorJson("Invalid avatar hash!"), http.StatusBadRequest)
		return
	} else if !slices.Contains(VALID_AVATAR_SIZES, r.URL.Query().Get("size")) {
		http.Error(w, errorJson("Invalid size parameter! Supported sizes: 256, 4096"),
			http.StatusBadRequest)
		return
	}
	// Retrieve avatar from the database
	// Assumptions: Every avatar is 1:1 aspect ratio, 4096x4096 max resolution, AVIF format
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
	// If ?size=256, downscale the avatar, otherwise for 4096, return the original lossless image
	data := avatar.Data
	// if size, err := strconv.Atoi(r.URL.Query().Get("size")); err == nil && size < 4096 {
	if r.URL.Query().Get("size") == "256" {
		size := 256
		// Decode original image
		originalImage, err := DecodeAVIF(avatar.Data)
		if err != nil {
			http.Error(w, errorJson("Failed to decode avatar image!"), http.StatusUnprocessableEntity)
			return
		}
		// Resize image
		resizedImage := originalImage
		if originalImage.Bounds().Dx() > size {
			resizedImage = imaging.Resize(originalImage, size, size, imaging.Lanczos)
			// Encode image
			data, err = EncodeAVIF(resizedImage, AVIF_QUALITY)
			if err != nil {
				handleInternalServerError(w, err)
				return
			}
		}
	}
	// Return the avatar
	w.Header().Set("Content-Type", "image/avif")
	http.ServeContent(w, r, avatar.Hash+".avif", avatar.CreatedAt, bytes.NewReader(data))
}

const MAX_AVATAR_SIZE_MB = 16
const MAX_AVATAR_SIZE = 1024 * 1024 * MAX_AVATAR_SIZE_MB
const MAX_AVATAR_RES = 4096

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
		if n, err := io.CopyN(avatarData, r.Body, MAX_AVATAR_SIZE); err != nil && !errors.Is(err, io.EOF) {
			handleInternalServerError(w, err)
			return
		} else if n == MAX_AVATAR_SIZE {
			http.Error(w,
				errorJson("Avatar data too large! Maximum size is "+strconv.Itoa(MAX_AVATAR_SIZE_MB)+" MB."),
				http.StatusBadRequest)
			return
		} else if n > 0 {
			// Decode the image
			originalImage, _, err := image.Decode(avatarData)
			if err != nil {
				http.Error(w, errorJson("Failed to decode avatar image! Supported formats: PNG, JPEG, GIF, WebP, BMP, TIFF"),
					http.StatusBadRequest)
				return
			}
			// Crop the image
			croppedImage := originalImage
			if originalImage.Bounds().Dx() != originalImage.Bounds().Dy() {
				res := min(originalImage.Bounds().Dx(), originalImage.Bounds().Dy(), MAX_AVATAR_RES)
				croppedImage = imaging.Fill(originalImage, res, res, imaging.Center, imaging.Lanczos)
			} else if originalImage.Bounds().Dx() > MAX_AVATAR_RES {
				croppedImage = imaging.Resize(originalImage, 4096, 4096, imaging.Lanczos)
			}
			// Encode the image
			data, err = EncodeAVIF(croppedImage, AVIF_QUALITY)
			if err != nil {
				handleInternalServerError(w, err)
				return
			}
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

	hashOrNil := &hash
	if hash == "" {
		hashOrNil = nil
	}
	propagateUserProfileUpdate(user.ID, struct {
		Avatar *string `json:"avatar"`
	}{Avatar: hashOrNil})
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

func propagateUserProfileUpdate(userID uuid.UUID, update interface{}) {
	if conns, ok := userConns.Load(userID); ok {
		rooms := make(map[string]struct{})
		conns.Range(func(_ chan<- interface{}, connInfo UserConnInfo) bool {
			rooms[connInfo.RoomID] = struct{}{}
			return true
		})
		for roomID := range rooms {
			if users, ok := roomMembers.Load(roomID); ok {
				users.Range(func(_ RoomConnID, conn chan<- interface{}) bool {
					conn <- UserProfileUpdateMessageOutgoing{
						Type: "user_profile_update",
						ID:   userID,
						Data: update,
					}
					return true
				})
			}
		}
	}
}
