package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
)

const version = "1.0.1"

/*
Endpoints:
- GET /
- POST /api/login
- POST /api/logout
- POST /api/register
- POST /api/forgot-password
- GET /api/forgot-password/:token
- POST /api/reset-password
- POST /api/change-password
- POST /api/change-username
- POST /api/change-email
- DELETE /api/delete-account
- GET /api/usernames?id=:id - Get usernames by ID, can accept multiple `id` query parameters
- POST /api/room - Create a new room
- GET /api/room/:id - Get the room's info
- PATCH /api/room/:id - Update the room's info
- WS /api/room/:id/join - Join an existing room
- GET /api/room/:id/subtitle - Get a subtitle from the room
- POST /api/room/:id/subtitle - Add a subtitle to the room

You can be a member of up to 3 rooms at once.
Rooms are deleted after 10 minutes of no members.
*/

var db *sql.DB
var config Config = Config{BasePath: "/", Port: 8000, Database: "postgres"}

type Config struct {
	Port          int    `json:"port"`
	BasePath      string `json:"basePath"`
	SecureCookies bool   `json:"secureCookies"`
	Database      string `json:"database"`
	DatabaseURL   string `json:"databaseUrl"`
	FrontendURL   string `json:"frontendUrl"`
	EmailSettings struct {
		Identity string `json:"identity"`
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
	} `json:"emailSettings"`
}

// TODO: implement e-mail verification option
func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version") {
		log.Println("concinnity version " + version)
		return
	}

	log.SetOutput(os.Stderr)
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Failed to read config file!", err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalln("Failed to parse config file!", err)
	}
	if config.Database == "mariadb" {
		config.Database = "mysql"
		dsn, err := mysql.ParseDSN(config.DatabaseURL)
		if err != nil {
			log.Fatalln("Failed to parse MariaDB DSN!", err)
		}
		dsn.MultiStatements = true
		dsn.ParseTime = true
		dsn.Params = map[string]string{"time_zone": "'+00:00'"} // dsn.Loc is already UTC
		config.DatabaseURL = dsn.FormatDSN()
	} else if config.Database != "postgres" {
		log.Fatalln("Unsupported database \"" + config.Database + "\" specified in config!")
	}
	db, err = sql.Open(config.Database, config.DatabaseURL)
	if err != nil {
		log.Fatalln("Failed to open connection to database!", err)
	}
	db.SetMaxOpenConns(10)
	CreateSqlTables()
	if slices.Contains(os.Args, "--upgrade") {
		UpgradeSqlTables()
	}
	PrepareSqlStatements()
	go PurgeExpiredDataTask()
	if !IsEmailConfigured() || config.FrontendURL == "" {
		log.Println("Note: Email settings and frontend URL are not configured for the forgot password functionality!")
	}

	// Endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != "GET" {
			http.NotFound(w, r)
		} else {
			StatusEndpoint(w, r)
		}
	})
	http.HandleFunc("POST /api/login", LoginEndpoint)
	http.HandleFunc("POST /api/logout", LogoutEndpoint)
	http.HandleFunc("POST /api/register", RegisterEndpoint)
	http.HandleFunc("POST /api/forgot-password", ForgotPasswordEndpoint)
	http.HandleFunc("GET /api/forgot-password/{token}", ForgotPasswordTokenEndpoint)
	http.HandleFunc("POST /api/reset-password", ResetPasswordEndpoint)
	http.HandleFunc("POST /api/change-password", ChangePasswordEndpoint)
	http.HandleFunc("POST /api/change-username", ChangeUsernameEndpoint)
	http.HandleFunc("POST /api/change-email", ChangeEmailEndpoint)
	http.HandleFunc("DELETE /api/delete-account", DeleteAccountEndpoint)
	http.HandleFunc("GET /api/usernames", GetUsernamesEndpoint)
	http.HandleFunc("GET /api/avatar/{hash}", GetAvatarEndpoint)
	http.HandleFunc("POST /api/room", CreateRoomEndpoint)
	http.HandleFunc("GET /api/room/{id}", GetRoomEndpoint)
	http.HandleFunc("PATCH /api/room/{id}", UpdateRoomEndpoint)
	http.HandleFunc("GET /api/room/{id}/join", JoinRoomEndpoint)
	http.HandleFunc("GET /api/room/{id}/subtitle", GetRoomSubtitleEndpoint)
	http.HandleFunc("POST /api/room/{id}/subtitle", CreateRoomSubtitleEndpoint)

	port := strconv.Itoa(config.Port)
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.SetOutput(os.Stdout)
	log.Println("Listening to port " + port)
	log.SetOutput(os.Stderr)
	log.Fatalln(http.ListenAndServe(":"+port, handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedOrigins([]string{"*"}), // Breaks credentialed auth
		handlers.AllowCredentials(),
	)(http.DefaultServeMux)))
}
