package main

import (
	"fmt"
	"log"
	"net/http"
	"pinterest/auth"
	"pinterest/pins"
	"pinterest/profile"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

func runServer(addr string) {
	r := mux.NewRouter()

	r.HandleFunc("/auth/signup", auth.HandleCreateUser).Methods("POST")
	r.HandleFunc("/auth/login", auth.HandleLoginUser).Methods("POST")
	r.HandleFunc("/auth/logout", auth.HandleLogoutUser).Methods("POST")

	r.HandleFunc("/profile/password", profile.HandleChangePassword).Methods("PUT")
	r.HandleFunc("/profile/edit", profile.HandleEditProfile).Methods("PUT")
	r.HandleFunc("/profile/delete", profile.HandleDeleteProfile).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", profile.HandleGetProfile).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", profile.HandleGetProfile).Methods("GET")

	pins := &pins.PinsStorage{
		Storage: pins.NewPinsSet(0),
	}

	r.HandleFunc("/pin", pins.Storage.AddPin).Methods("POST")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.GetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.DelPinByID).Methods("DELETE")

	r.HandleFunc("/board/", boardHandler) // Will split later

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://52.59.228.167:8080"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	fmt.Printf("Starting server at localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func main() {
	runServer(":8080")
}
