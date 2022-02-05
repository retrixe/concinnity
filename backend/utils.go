package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func errorJson(err string) string {
	json, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: err})
	return string(json)
}

func handleInternalServerError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, errorJson("Internal Server Error!"), http.StatusInternalServerError)
}
