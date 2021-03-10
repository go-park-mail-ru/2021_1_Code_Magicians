package auth

import (
	"encoding/json"
	"io/ioutil"
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

// UserInput if used to parse JSON with users' data
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

type cookieInfo struct {
	userID int
	cookie *http.Cookie
}

type sessionMap struct {
	sessions map[string]cookieInfo // key is cookie value, for easier lookup
	mu       sync.Mutex
}

var users usersMap = usersMap{users: make(map[int]user), lastFreeUserID: 0}
var sessions sessionMap = sessionMap{sessions: make(map[string]cookieInfo)}

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

// HandleCreateUser creates user with parameters passed in JSON
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	userInput := new(UserInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users.mu.Lock()

	// Checking for username uniqueness
	for _, user := range users.users {
		if user.username == userInput.Username {
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

	w.WriteHeader(http.StatusCreated)
}

// checkCookies returns *cookieInfo and true if cookie is present in sessions slice, nil and false othervise
func checkCookies(r *http.Request) (*cookieInfo, bool) {
	cookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, false
	}

	sessions.mu.Lock()
	userCookieInfo, ok := sessions.sessions[cookie.Value]
	sessions.mu.Unlock()

	if !ok { // cookie was not found
		return nil, false
	}

	return &userCookieInfo, true
}

// HandleLoginUser logs user in using provided username and password
func HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	userInput := new(UserInput)
	err := decoder.Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, cookieFound := checkCookies(r)
	if cookieFound {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for id, user := range users.users {
		if user.username == userInput.Username {
			if user.password != userInput.Password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sessionValue := randSeq(cookieLength) // cookie value - random string
			expiration := time.Now().Add(expirationTime)
			cookie := http.Cookie{
				Name:     "session_id",
				Value:    sessionValue,
				Expires:  expiration,
				HttpOnly: true, // So that frontend won't have direct access to cookies
			}
			http.SetCookie(w, &cookie)

			sessions.mu.Lock()
			sessions.sessions[sessionValue] = cookieInfo{id, &cookie}
			sessions.mu.Unlock()

			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
	return
}

// HandleLogoutUser tries to log user out of current session
func HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookieInfo, found := checkCookies(r)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userCookieInfo.cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookieInfo.cookie)

	cookieValue := userCookieInfo.cookie.Value
	sessions.mu.Lock()
	delete(sessions.sessions, cookieValue)
	sessions.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	return
}
