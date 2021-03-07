package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type user struct {
	username  string
	password  string // TODO: hashing
	firstName string
	lastName  string
	avatar    string // path to avatar
}

type UserInput struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
}

type usersMap struct {
	users          map[int]user
	lastFreeUserID int
	mu             sync.Mutex
}

var users usersMap
var sessions map[int]bool

func (users *usersMap) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	newUserInput := new(UserInput)
	err := decoder.Decode(newUserInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.Write([]byte(`{"code": 400}`))
		return
	}

	users.mu.Lock()

	// TODO: Check if these fields are empty
	users.users[users.lastFreeUserID] = user{
		username:  newUserInput.Username,
		password:  newUserInput.Password,
		firstName: newUserInput.FirstName,
		lastName:  newUserInput.LastName,
		avatar:    newUserInput.Avatar,
	}
	users.lastFreeUserID++

	users.mu.Unlock()

	w.Write([]byte(`{"code": 201}`)) // returning success code
}

func (users *usersMap) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	newUserInput := new(UserInput)
	err := decoder.Decode(newUserInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.Write([]byte(`{"code": 400}`))
		return
	}

	for id, user := range users.users {
		if user.username == newUserInput.Username {
			if user.password != newUserInput.Password {
				w.Write([]byte(`{"code": 400}`))
				return
			}
			sessions[id] = true
			w.Write([]byte(`{"code": 200, "X-Expires-After": "Expires: Mon, 29 Mar 2021 10:00:00 GMT"}`)) // TODO: normal datetime
			return
		}
	}

	w.Write([]byte(`{"code": 400}`)) // No users with passed username
	return
}

func (users *usersMap) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	// TODO: "Current session" handling

	//logout all users
	for id := range sessions {
		delete(sessions, id)
	}

	w.Write([]byte(`{"code": 200}`)) // No users with passed username
	return
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println(r.URL.Path)
	switch r.URL.Path {
	case "/auth/signup":
		if r.Method != http.MethodPost {
			w.Write([]byte(`{"code": 400}`))
			return
		}
		users.HandleCreateUser(w, r)

	case "/auth/login":
		if r.Method != http.MethodGet {
			w.Write([]byte(`{"code": 400}`))
			return
		}
		users.HandleLoginUser(w, r)

	case "/auth/logout":
		if r.Method != http.MethodGet {
			w.Write([]byte(`{"code": 400}`))
			return
		}
		users.HandleLogoutUser(w, r)

	default:
		w.Write([]byte(`{"code": 400}`))
		return
	}
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
	users = usersMap{users: make(map[int]user), lastFreeUserID: 0}
	sessions = make(map[int]bool)
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	mux.HandleFunc("/auth/", authHandler)
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
