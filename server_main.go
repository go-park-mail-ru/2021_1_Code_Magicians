package main

import (
	"fmt"
	"net/http"
)

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /auth handling
}

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
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	mux.HandleFunc("/auth/", authHandler)

	mux.HandleFunc("/pin/", pinHandler)
	mux.HandleFunc("/pins/", pinHandler)

	mux.HandleFunc("/board/", boardHandler)
	mux.HandleFunc("/profile/", profileHandler)

	fmt.Printf("Starting server at localhost%s\n", addr)
	server.ListenAndServe()
}

func main() {
	runServer(":8080")
}
