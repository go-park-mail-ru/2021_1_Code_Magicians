package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
}

type usersMap struct {
	users          map[int]user
	lastFreeUserID int
	mu             sync.Mutex
}

type sessionMap struct {
	sessions map[string]int // key - cookie value, value - user's id
	mu       sync.Mutex
}

var users usersMap = usersMap{users: make(map[int]user), lastFreeUserID: 0}
var sessions sessionMap = sessionMap{sessions: make(map[string]int)}

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

	// TODO: Check if these fields are empty and for login uniqueness
	users.users[users.lastFreeUserID] = user{
		username:  newUserInput.Username,
		password:  newUserInput.Password,
		firstName: newUserInput.FirstName,
		lastName:  newUserInput.LastName,
		avatar:    newUserInput.Avatar,
	}
	users.lastFreeUserID++

	users.mu.Unlock()

	log.Printf("Created user %s successfully", newUserInput.Username)
	w.Write([]byte(`{"code": 201}`)) // returning success code
}

// checkCookies returns users' id, cookie value and true if cookie is present in sessions slice, -1, "" and false othervise
func checkCookies(r *http.Request) (int, string, bool) {
	cookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return -1, "", false
	}

	sessions.mu.Lock()
	id, ok := sessions.sessions[cookie.Value]
	sessions.mu.Unlock()

	if !ok { // cookie was not found
		return -1, "", false
	}

	return id, cookie.Value, true
}

func (users *usersMap) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	userInput := new(UserInput)
	err := decoder.Decode(userInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.Write([]byte(`{"code": 400}`))
		return
	}

	_, _, cookieFound := checkCookies(r)
	if cookieFound {
		log.Printf("Cannot log in: user %s is already logged in", userInput.Username)
		w.Write([]byte(`{"code": 400}`))
		return
	}

	for id, user := range users.users {
		if user.username == userInput.Username {
			if user.password != userInput.Password {
				log.Printf("Password %s does not match", userInput.Password)
				w.Write([]byte(`{"code": 400}`))
				return
			}

			sessions.mu.Lock()
			sessions.sessions[strconv.Itoa(id)] = id
			sessions.mu.Unlock()

			expiration := time.Now().Add(10 * time.Hour)
			cookie := http.Cookie{
				Name:     "session_id",
				Value:    strconv.Itoa(id), // TODO: replace with random string
				Expires:  expiration,
				HttpOnly: true, // So that frontend won't have direct access to cookies
			}
			http.SetCookie(w, &cookie)

			log.Printf("Logged in user %s successfully", userInput.Username)
			w.Write([]byte(`{"code": 200}`))
			return
		}
	}

	log.Printf("User %s not found", userInput.Username)
	w.Write([]byte(`{"code": 400}`)) // No users with supplied username
	return
}

func (users *usersMap) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	id, cookieValue, found := checkCookies(r)
	if !found {
		log.Print("No cookies passed - user is not logged in")
		w.Write([]byte(`{"code": 400}`))
		return
	}

	sessions.mu.Lock()
	delete(sessions.sessions, cookieValue)
	sessions.mu.Unlock()

	users.mu.Lock()
	log.Printf("Successfully logged out user: %s", users.users[id].username)
	users.mu.Unlock()
	w.Write([]byte(`{"code": 200}`))
	return
}

// Handler handles responses that require authentification
func Handler(w http.ResponseWriter, r *http.Request) {
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
