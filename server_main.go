package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"pinterest/pins"
)

type PinsStorage struct {
	storage *pins.UserPinSet
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /auth handling
}

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

func runServer(addr string) {
	mux := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	pins := &PinsStorage{
		storage: pins.NewPinsSet(0),
	}

	mux.HandleFunc("/auth/", authHandler)

	mux.HandleFunc("/pin/", pins.storage.AddPin).Methods("POST")
	mux.HandleFunc("/pins/{id:[0-9]+}", pins.storage.GetPinByID).Methods("GET")
	mux.HandleFunc("/pins/{id:[0-9]+}", pins.storage.DelPinByID).Methods("DELETE")

	mux.HandleFunc("/board/", boardHandler)
	mux.HandleFunc("/profile/", profileHandler)

	fmt.Printf("Starting server at localhost%s\n", addr)
	server.ListenAndServe()
}

func main() {
	runServer(":8080")
}
