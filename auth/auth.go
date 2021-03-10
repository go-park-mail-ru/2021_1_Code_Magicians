package auth

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

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
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

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

// searchUser returns user's id and true if user is found, -1 and false otherwise
func searchUser(username string, password string) (int, bool) {
	for id, user := range users.users {
		if user.username == username {
			if user.password == password {
				return id, true
			}

			break
		}
	}
	return -1, false
}

// HandleLoginUser logs user in using provided username and password
func HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(UserInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, cookieFound := checkCookies(r)
	if cookieFound { // User is already logged in
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, exists := searchUser(userInput.Username, userInput.Password)

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
	}
	http.SetCookie(w, &cookie)

	sessions.mu.Lock()
	sessions.sessions[sessionValue] = cookieInfo{id, &cookie}
	sessions.mu.Unlock()

	w.WriteHeader(http.StatusOK)
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
