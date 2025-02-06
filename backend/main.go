package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
)

/*
Endpoints:
- GET /
- POST /api/login
- POST /api/logout
- POST /api/register
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
	// Note:
	// - MySQL support is best effort and untested! PostgreSQL is the only recommended option.
	// - MySQL requires ?parseTime=true&multiStatements=true to be set on the URL
	// - MySQL should be using the UTC time zone, anything else may or may not cause issues
	Database    string `json:"database"`
	DatabaseURL string `json:"databaseUrl"`
}

// TODO: implement e-mail verification option, add forgot password endpoint, room member limit
func main() {
	log.SetOutput(os.Stderr)
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Failed to read config file!", err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalln("Failed to parse config file!", err)
	}
	db, err = sql.Open(config.Database, config.DatabaseURL)
	if err != nil {
		log.Fatalln("Failed to open connection to database!", err)
	}
	db.SetMaxOpenConns(10)
	CreateSqlTables()
	PrepareSqlStatements()
	go CleanInactiveRoomsTask()

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
	http.HandleFunc("GET /api/usernames", GetUsernamesEndpoint)
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
