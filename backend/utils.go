package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/crypto/argon2"
)

func errorJson(err string) string {
	json, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: err})
	return string(json)
}

func handleInternalServerError(w http.ResponseWriter, err error) {
	log.Println("Internal Server Error!", err)
	http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
}

func wsInternalError(c *websocket.Conn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	log.Println("Internal Server Error!", err)
	_ = c.Write(ctx, websocket.MessageText, []byte(errorJson("Internal Server Error!")))
	_ = c.Close(websocket.StatusInternalError, "Internal Server Error!")
}

func wsError(c *websocket.Conn, err string, code websocket.StatusCode) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_ = c.Write(ctx, websocket.MessageText, []byte(errorJson(err)))
	_ = c.Close(code, err)
}

func GetTokenFromHTTP(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if cookie, err := r.Cookie("token"); err == nil {
		token = cookie.Value
	}
	return token
}

// GenerateSalt returns a 16-character salt readable in UTF-8 format as well.
func GenerateSalt() []byte {
	saltBytes := make([]byte, 12)
	_, _ = rand.Read(saltBytes)
	salt := base64.RawStdEncoding.EncodeToString(saltBytes)
	return []byte(salt)
}

func HashPassword(password string, salt []byte) string {
	params := "$argon2id$v=19$m=51200,t=1,p=4$" // Currently fixed only.
	key := argon2.IDKey([]byte(password), salt, 1, 51200, 4, 32)
	return params + base64.RawStdEncoding.EncodeToString(salt) +
		"$" + base64.RawStdEncoding.EncodeToString(key)
}

func ComparePassword(password string, hash string) bool {
	encodeSplit := strings.Split(hash, "$")
	salt, _ := base64.RawStdEncoding.DecodeString(encodeSplit[len(encodeSplit)-2])
	key := argon2.IDKey([]byte(password), salt, 1, 51200, 4, 32)
	hashValue := encodeSplit[len(encodeSplit)-1]
	return hashValue == base64.RawStdEncoding.EncodeToString(key)
}

func IsEmailConfigured() bool {
	return config.EmailSettings.Username != "" &&
		config.EmailSettings.Password != "" &&
		config.EmailSettings.Host != ""
}

func SendHTMLEmail(email string, subject string, body string) error {
	auth := smtp.PlainAuth(
		config.EmailSettings.Identity,
		config.EmailSettings.Username,
		config.EmailSettings.Password,
		config.EmailSettings.Host)
	from := config.EmailSettings.Identity
	if from == "" {
		from = config.EmailSettings.Username
	}
	host := config.EmailSettings.Host
	if !strings.Contains(host, ":") {
		host += ":587"
	}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		strings.ReplaceAll(body, "\n", "\r\n") + "\r\n")
	return smtp.SendMail(host, auth, from, []string{email}, msg)
}

func DecodeAVIF(data []byte) (image.Image, error) {
	// Create a temporary file containing the AVIF data
	file, err := os.CreateTemp(os.TempDir(), "concinnity-*.avif")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		return nil, err
	} else if err := file.Close(); err != nil {
		return nil, err
	}

	// Decode to PNG and read back the data
	defer os.Remove(file.Name() + ".png")
	if err := exec.Command("avifdec", "--png-compress", "0", file.Name(), file.Name()+".png").Run(); err != nil {
		return nil, err
	}
	file, err = os.Open(file.Name() + ".png")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func EncodeAVIF(image image.Image) ([]byte, error) {
	// Create a temporary file to encode the image to PNG
	file, err := os.CreateTemp(os.TempDir(), "concinnity-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(file, image); err != nil {
		return nil, err
	} else if err := file.Close(); err != nil {
		return nil, err
	}

	// Encode to AVIF using avifenc
	defer os.Remove(file.Name() + ".avif")
	// Anecdotal samples:
	// - Quality wise, 80 can be visually lacking with some images, 85 is fine, 90 is almost perfect
	// - File size wise, 85 sits midway between 80-90 for similar quality to 90
	if err := exec.Command("avifenc", "-q", "85", file.Name(), file.Name()+".avif").Run(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(file.Name() + ".avif")
	if err != nil {
		return nil, err
	}
	return data, nil
}
