package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
)

/*
Endpoints:
- GET /
- POST /api/login
- POST /api/logout
- POST /api/register
- GET /api/room/:id - Get the room's info
- POST /api/room - Create a new room and join it
- POST /api/room/:id - Join an existing room
- WS /api/room/:id - Get live updates to room's info
- GET /api/room/:id/leave - Leave a room

You can be a member of up to 3 rooms at once.
Rooms are deleted after 10 minutes of no members.
Implement a rate limit of 3reqs/10min on creating rooms.
*/

var db *sql.DB
var secureCookies bool

// TODO: implement e-mail verification option, add forgot password endpoint, room member limit
func main() {
	log.SetOutput(os.Stderr)
	// TODO: use environment variables or config
	secureCookies = false
	connStr := "dbname=concinnity user=postgres host=localhost password=postgres sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Panicln("Failed to open connection to database!", err)
	}
	db.SetMaxOpenConns(10)
	CreateSqlTables()
	PrepareSqlStatements()

	// TODO: use gin or iris or httprouter maybe
	// Endpoints
	http.HandleFunc("/", StatusEndpoint)
	http.HandleFunc("/api/login", LoginEndpoint)
	http.HandleFunc("/api/logout", LogoutEndpoint)
	http.HandleFunc("/api/register", RegisterEndpoint)
	// TODO handle this in another way
	// No URL params
	http.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		println(r.Method)
		if r.Method == "POST" {
			// POST /api/room
			CreateRoom(w, r)
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
	})
	// A / in the end means there's URL params
	http.HandleFunc("/api/room/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// GET /api/room/:id and GET /api/room/:id/leave
			GetRoom(w, r)
		} else if r.Method == "POST" {
			// POST /api/room/:id
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else if r.Method == "OPTIONS" {
			// OPTIONS /api/room/:id
			http.Error(w, errorJson("Not Implemented!"), http.StatusNotImplemented) // TODO
		} else {
			http.Error(w, errorJson("Method Not Allowed!"), http.StatusMethodNotAllowed)
		}
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
