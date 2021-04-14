package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"pinterest/interfaces/middleware"
	"time"
)

// AuthInfo keep information about apps and cookies needed for auth package
type AuthInfo struct {
	userApp   application.UserAppInterface
	cookieApp application.CookieAppInterface
	s3App     application.S3AppInterface
	boardApp  application.BoardAppInterface // For initial user's board
}

func NewAuthInfo(userApp application.UserAppInterface, cookieApp application.CookieAppInterface,
	s3App application.S3AppInterface, boardApp application.BoardAppInterface) *AuthInfo {
	return &AuthInfo{
		userApp:   userApp,
		cookieApp: cookieApp,
		s3App:     s3App,
		boardApp:  boardApp,
	}
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

	valid, _ := userInput.Validate()
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

	user, _ := info.userApp.GetUserByUsername(newUser.Username)
	if user != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	newUser.UserID, err = info.userApp.CreateUser(&newUser)
	if err != nil {
		if err.Error() == "Username or email is already taken" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie, err := info.cookieApp.GenerateCookie()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)

		info.userApp.DeleteUser(newUser.UserID)
		return
	}

	err = info.cookieApp.AddCookie(&entity.CookieInfo{newUser.UserID, cookie})
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

	user, err := info.userApp.CheckUserCredentials(userInput.Username, userInput.Password)

	if err != nil {
		switch err.Error() {
		case "Password does not match":
			w.WriteHeader(http.StatusUnauthorized)
		case "No user found with such username":
			w.WriteHeader(http.StatusNotFound)
		default:
			{
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	cookie, err := info.cookieApp.GenerateCookie()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = info.cookieApp.AddCookie(&entity.CookieInfo{user.UserID, cookie})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
	return
}

// HandleLogoutUser logs current user out of their session
func (info *AuthInfo) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)

	err := info.cookieApp.RemoveCookie(userCookie)
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
	_, found := middleware.CheckCookies(r, info.cookieApp)
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
