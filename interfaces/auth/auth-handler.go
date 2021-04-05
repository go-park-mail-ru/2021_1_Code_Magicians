package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"pinterest/domain/entity"
	"time"
)

// Users is a map of all existing users
var Users entity.UsersMap = entity.UsersMap{Users: make(map[int]entity.User), LastFreeUserID: 0}
var sessions entity.SessionMap = entity.SessionMap{Sessions: make(map[string]entity.CookieInfo)}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// generateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func generateCookie(length int, duration time.Duration) (*http.Cookie, error) {
	sessionValue, err := generateRandomString(cookieLength) // cookie value - random string
	if err != nil {
		return nil, err
	}

	expiration := time.Now().Add(duration)
	return &http.Cookie{
		Name:     "session_id",
		Value:    sessionValue,
		Expires:  expiration,
		HttpOnly: true, // So that frontend won't have direct access to cookies
		Path:     "/",  // Cookie should be usable on entire website
	}, nil
}

const cookieLength int = 40
const expirationTime time.Duration = 10 * time.Hour

// HandleCreateUser creates user with parameters passed in JSON
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userInput := new(entity.UserRegInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, err := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newUser entity.User
	err = newUser.UpdateFrom(userInput)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, alreadyExists := FindUser(newUser.Username)
	if alreadyExists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if newUser.Avatar == "" {
		newUser.Avatar = "/assets/img/default-avatar.jpg" // default user avatar path
	}

	Users.Mu.Lock()

	id := Users.LastFreeUserID

	Users.Users[id] = newUser
	Users.LastFreeUserID++

	Users.Mu.Unlock()

	cookie, err := generateCookie(cookieLength, expirationTime)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)

		// Removing user we just created
		Users.Mu.Lock()
		delete(Users.Users, id)
		Users.Mu.Unlock()
		return
	}

	http.SetCookie(w, cookie)

	sessions.Mu.Lock()
	sessions.Sessions[cookie.Value] = entity.CookieInfo{id, cookie}
	sessions.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// CheckCookies returns *CookieInfo and true if cookie is present in sessions slice, nil and false othervise
func CheckCookies(r *http.Request) (*entity.CookieInfo, bool) {
	cookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, false
	}

	sessions.Mu.Lock()
	userCookieInfo, ok := sessions.Sessions[cookie.Value]
	sessions.Mu.Unlock()

	if !ok { // cookie was not found
		return nil, false
	}

	if userCookieInfo.Cookie.Expires.Before(time.Now()) {
		sessions.Mu.Lock()
		delete(sessions.Sessions, cookie.Value)
		sessions.Mu.Unlock()
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

	userInput := new(entity.UserLoginInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, exists := checkUserCredentials(userInput.Username, userInput.Password)

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cookie, err := generateCookie(cookieLength, expirationTime)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)

	sessions.Mu.Lock()
	sessions.Sessions[cookie.Value] = entity.CookieInfo{id, cookie}
	sessions.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
	return
}

// HandleLogoutUser tries to log user out of current session
func HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookie, _ := CheckCookies(r)
	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	cookieValue := userCookie.Cookie.Value
	sessions.Mu.Lock()
	delete(sessions.Sessions, cookieValue)
	sessions.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
	return
}

// HandleCheckUser checks if current user is logged in
func HandleCheckUser(w http.ResponseWriter, r *http.Request) {
	_, found := CheckCookies(r)
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
