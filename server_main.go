package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"pinterest/pins"
  "pinterest/auth"
  "pinterest/profile"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

func runServer(addr string) {
	r := mux.NewRouter()

	r.HandleFunc("/auth/signup", auth.HandleCreateUser).Methods("POST")
	r.HandleFunc("/auth/login", auth.HandleLoginUser).Methods("GET")
	r.HandleFunc("/auth/logout", auth.HandleLogoutUser).Methods("GET")

	r.HandleFunc("/profile/change-password", profile.HandleChangePassword).Methods("POST")
	r.HandleFunc("/profile/edit", profile.HandleEditProfile).Methods("PUT")
	r.HandleFunc("/profile/delete", profile.HandleDeleteProfile).Methods("DELETE")
	r.HandleFunc("/profile/{username}", profile.HandleGetProfile).Methods("GET")

	pins := &PinsStorage{
		storage: pins.NewPinsSet(0),
	}

	r.HandleFunc("/pin/", pins.storage.AddPin).Methods("POST")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.storage.GetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.storage.DelPinByID).Methods("DELETE")

	r.HandleFunc("/board/", boardHandler) // Will split later

	fmt.Printf("Starting server at localhost%s\n", addr)
	http.ListenAndServe(addr, r)
}

func main() {
	runServer(":8080")
}
