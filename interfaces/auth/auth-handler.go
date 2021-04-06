package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	application "pinterest/applicaton"
	"pinterest/domain/entity"
	"pinterest/interfaces/middleware"
	"time"
)

type AuthInfo struct {
	UserApp      application.UserAppInterface
	CookieApp    application.CookieAppInterface
	CookieLength int
	Duration     time.Duration
}

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

func (info *AuthInfo) generateCookie() (*http.Cookie, error) {
	sessionValue, err := generateRandomString(info.CookieLength) // cookie value - random string
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(info.Duration)
	return &http.Cookie{
		Name:     "session_id",
		Value:    sessionValue,
		Expires:  expirationTime,
		HttpOnly: true, // So that frontend won't have direct access to cookies
		Path:     "/",  // Cookie should be usable on entire website
	}, nil
}

// HandleCreateUser creates user with parameters passed in JSON
func (info *AuthInfo) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
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

	user, _ := info.UserApp.GetUserByUsername(newUser.Username)
	if user != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if newUser.Avatar == "" {
		newUser.Avatar = "/assets/img/default-avatar.jpg" // default user avatar path
	}

	newUser.UserID, err = info.UserApp.SaveUser(&newUser)
	if err != nil {
		if err.Error() == "Username or email is already taken" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie, err := info.generateCookie()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)

		// TODO: delete user
		return
	}

	err = info.CookieApp.AddCookie(&entity.CookieInfo{newUser.UserID, cookie})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusCreated)
}

// HandleLoginUser logs user in using provided username and password
func (info *AuthInfo) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userInput := new(entity.UserLoginInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := info.UserApp.CheckUserCredentials(userInput.Username, userInput.Password)

	if err != nil {
		switch err.Error() {
		case "Password does not match":
			w.WriteHeader(http.StatusUnauthorized)
		case "No user found with such username":
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	cookie, err := info.generateCookie()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = info.CookieApp.AddCookie(&entity.CookieInfo{id, cookie})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
	return
}

// HandleLogoutUser tries to log user out of current session
func (info *AuthInfo) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookie := r.Context().Value("cookieInfo").(*entity.CookieInfo)

	err := info.CookieApp.RemoveCookie(userCookie)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	w.WriteHeader(http.StatusNoContent)
	return
}

// HandleCheckUser checks if current user is logged in
func (info *AuthInfo) HandleCheckUser(w http.ResponseWriter, r *http.Request) {
	_, found := middleware.CheckCookies(r, info.CookieApp)
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
