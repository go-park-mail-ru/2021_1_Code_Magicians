package auth

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

// Users is a map of all existing users
var Users UsersMap = UsersMap{Users: make(map[int]User), LastFreeUserID: 0}
var sessions sessionMap = sessionMap{sessions: make(map[string]CookieInfo)}

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
	defer r.Body.Close()

	userInput := new(UserIO)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userInput.Username == "" || userInput.Password == "" ||
		userInput.FirstName == "" || userInput.LastName == "" ||
		userInput.Email == "" || userInput.Avatar == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Users.Mu.Lock()

	// Checking for username uniqueness
	for _, user := range Users.Users {
		if user.Username == userInput.Username {
			Users.Mu.Unlock()
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	Users.Users[Users.LastFreeUserID] = User{
		Username:  userInput.Username,
		Password:  userInput.Password,
		FirstName: userInput.FirstName,
		LastName:  userInput.LastName,
		Email:     userInput.Email,
		Avatar:    userInput.Avatar,
	}
	Users.LastFreeUserID++

	Users.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// CheckCookies returns *CookieInfo and true if cookie is present in sessions slice, nil and false othervise
func CheckCookies(r *http.Request) (*CookieInfo, bool) {
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

// FindUser tries to find user with passed username in Users map
func FindUser(username string) (int, bool) {
	Users.Mu.Lock()
	defer Users.Mu.Unlock()
	for id, user := range Users.Users {
		if user.Username == username {
			return id, true
		}
	}
	return -1, false
}

// checkUserCredentials returns user's id and true if user credentials match, -1 and false otherwise
func checkUserCredentials(username string, password string) (int, bool) {
	id, found := FindUser(username)
	if !found {
		return -1, false
	}

	Users.Mu.Lock()
	defer Users.Mu.Unlock()
	if Users.Users[id].Password == password {
		return id, true
	}

	return -1, false
}

// HandleLoginUser logs user in using provided username and password
func HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userInput := new(UserIO)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userInput.Username == "" || userInput.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, cookieFound := CheckCookies(r)
	if cookieFound { // User is already logged in
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, exists := checkUserCredentials(userInput.Username, userInput.Password)

	if !exists {
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
		Path:     "/",  // Cookie should be usable on entire website
	}
	http.SetCookie(w, &cookie)

	sessions.mu.Lock()
	sessions.sessions[sessionValue] = CookieInfo{id, &cookie}
	sessions.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	return
}

// HandleLogoutUser tries to log user out of current session
func HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookieInfo, found := CheckCookies(r)
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
