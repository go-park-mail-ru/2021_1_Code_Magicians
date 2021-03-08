package auth

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randSeq generates random string with length of n
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const cookieLength int = 30
const expirationTime time.Duration = 10 * time.Hour

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userInput := new(UserInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users.mu.Lock()

	// Checking for username uniqueness
	for _, user := range users.users {
		if user.username == userInput.Username {
			log.Printf("Username is already taken: %s", userInput.Username)
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	// TODO: Check if these fields are empty
	users.users[users.lastFreeUserID] = user{
		username:  userInput.Username,
		password:  userInput.Password,
		firstName: userInput.FirstName,
		lastName:  userInput.LastName,
		avatar:    userInput.Avatar,
	}
	users.lastFreeUserID++

	users.mu.Unlock()

	log.Printf("Created user %s successfully", userInput.Username)
	w.WriteHeader(http.StatusCreated)
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

func HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	userInput := new(UserInput)
	err := decoder.Decode(userInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _, cookieFound := checkCookies(r)
	if cookieFound {
		log.Printf("Cannot log in: user %s is already logged in", userInput.Username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for id, user := range users.users {
		if user.username == userInput.Username {
			if user.password != userInput.Password {
				log.Printf("Password %s does not match", userInput.Password)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sessionValue := randSeq(cookieLength) // cookie value - random string
			sessions.mu.Lock()
			sessions.sessions[sessionValue] = id
			sessions.mu.Unlock()

			expiration := time.Now().Add(expirationTime)
			cookie := http.Cookie{
				Name:     "session_id",
				Value:    sessionValue,
				Expires:  expiration,
				HttpOnly: true, // So that frontend won't have direct access to cookies
			}
			http.SetCookie(w, &cookie)

			log.Printf("Logged in user %s successfully", userInput.Username)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	log.Printf("User %s not found", userInput.Username)
	w.WriteHeader(http.StatusUnauthorized)
	return
}

func HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	id, cookieValue, found := checkCookies(r)
	if !found {
		log.Print("No cookies passed - user is not logged in or cookie is inactive")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessions.mu.Lock()
	delete(sessions.sessions, cookieValue)
	sessions.mu.Unlock()

	users.mu.Lock()
	log.Printf("Successfully logged out user: %s", users.users[id].username)
	users.mu.Unlock()
	w.WriteHeader(http.StatusOK)
	return
}
