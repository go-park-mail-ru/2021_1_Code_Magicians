package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"pinterest/pins"
  "pinterest/auth"
)


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
	r := mux.NewRouter()

	r.HandleFunc("/auth/signup", auth.HandleCreateUser).Methods("POST")
	r.HandleFunc("/auth/login", auth.HandleLoginUser).Methods("GET")
	r.HandleFunc("/auth/logout", auth.HandleLogoutUser).Methods("GET")

	pins := &pins.PinsStorage{
		Storage: pins.NewPinsSet(0),
	}

	r.HandleFunc("/auth/", authHandler)

	r.HandleFunc("/pin/", pins.Storage.AddPin).Methods("POST")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.GetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.DelPinByID).Methods("DELETE")
	r.HandleFunc("/profile/", profileHandler) // Will split later

	r.HandleFunc("/board/", boardHandler) // Will split later

	fmt.Printf("Starting server at localhost%s\n", addr)
	http.ListenAndServe(addr, r)
}

func main() {
	runServer(":8080")
}
