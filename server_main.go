package main

import (
	"fmt"
	"net/http"
	"pinterest/auth"
	"pinterest/profile"

	"github.com/gorilla/mux"
)

func pinHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
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

	r.HandleFunc("/profile/change-password", profile.HandleChangePassword).Methods("POST")
	r.HandleFunc("/profile/edit", profile.HandleEditProfile).Methods("PUT")
	r.HandleFunc("/profile/delete", profile.HandleDeleteProfile).Methods("DELETE")
	r.HandleFunc("/profile/{username}", profile.HandleGetProfile).Methods("GET")

	r.HandleFunc("/pin/", pinHandler)  // Will split later
	r.HandleFunc("/pins/", pinHandler) // Will split later

	r.HandleFunc("/board/", boardHandler) // Will split later

	fmt.Printf("Starting server at localhost%s\n", addr)
	http.ListenAndServe(addr, r)
}

func main() {
	runServer(":8080")
}
