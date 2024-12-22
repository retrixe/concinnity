package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
	"golang.org/x/net/websocket"
)

/*
Endpoints:
- GET /
- POST /api/login
- POST /api/logout
- POST /api/register
- POST /api/room - Create a new room
- GET /api/room/:id - Get the room's info
- PATCH /api/room/:id - Update the room's info
- WS /api/room/:id/join - Join an existing room

You can be a member of up to 3 rooms at once.
Rooms are deleted after 10 minutes of no members.
*/

var db *sql.DB
var config Config

type Config struct {
	SecureCookies bool   `json:"secureCookies"`
	DatabaseURL   string `json:"databaseUrl"`
}

// TODO: implement e-mail verification option, add forgot password endpoint, room member limit
func main() {
	log.SetOutput(os.Stderr)
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Panicln("Failed to read config file!", err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Panicln("Failed to parse config file!", err)
	}
	db, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Panicln("Failed to open connection to database!", err)
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
	http.HandleFunc("POST /api/room", CreateRoomEndpoint)
	http.HandleFunc("GET /api/room/{id}", GetRoomEndpoint)
	http.HandleFunc("PATCH /api/room/{id}", UpdateRoomEndpoint)
	http.HandleFunc("GET /api/room/{id}/join", func(w http.ResponseWriter, r *http.Request) {
		s := websocket.Server{Handler: JoinRoomEndpoint, Handshake: JoinRoomEndpointHandshake}
		s.ServeHTTP(w, r)
	})

	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.SetOutput(os.Stdout)
	log.Println("Listening to port " + port)
	log.SetOutput(os.Stderr)
	log.Fatalln(http.ListenAndServe(":"+port, handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authentication"}),
		handlers.AllowedOrigins([]string{"*"}), // Breaks credentialed auth
		handlers.AllowCredentials(),
	)(http.DefaultServeMux)))
}
