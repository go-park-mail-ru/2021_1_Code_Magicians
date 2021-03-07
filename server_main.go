package main

import (
	"fmt"
	"net/http"
	"pinterest/auth"
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
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	mux.HandleFunc("/auth/", auth.Handler)
	mux.HandleFunc("/profile/", profileHandler)

	mux.HandleFunc("/pin/", pinHandler)
	mux.HandleFunc("/pins/", pinHandler)

	mux.HandleFunc("/board/", boardHandler)

	fmt.Printf("Starting server at localhost%s\n", addr)
	server.ListenAndServe()
}

func main() {
	runServer(":8080")
}
