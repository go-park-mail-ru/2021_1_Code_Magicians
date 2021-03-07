package main

import (
	"fmt"
	"net/http"
	"pinterest/pins"
)

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /auth handling
}

func pinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println(r.URL.Path)
	switch r.URL.Path {

	case "/pins/{id:[0-9]+}":
		if r.Method == http.MethodGet {
			pins.MyPins.GetPinByID(w, r)
		} else if r.Method == http.MethodDelete {
			pins.MyPins.DelPinByID(w, r)
		} else {
			w.Write([]byte(`{"code": 400}`))
			return
		}
	case "/pin":
		if r.Method != http.MethodPost {
			w.Write([]byte(`{"code": 400}`))
			return
		}
		pins.MyPins.AddPin(w, r)
	default:
		w.Write([]byte(`{"code": 400}`))
		return
	}
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
