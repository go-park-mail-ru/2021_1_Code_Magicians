package profile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ProfileInfo keep information about apps and cookies needed for profile package
type ProfileInfo struct {
	UserApp   application.UserAppInterface
	CookieApp application.CookieAppInterface
	S3App     application.S3AppInterface
}

//HandleChangePassword changes password of current user
func (profileInfo *ProfileInfo) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserPassChangeInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, err := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := profileInfo.UserApp.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // TODO: error handling
		return
	}

	user.Password = userInput.Password
	err = profileInfo.UserApp.SaveUser(user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleEditProfile edits profile of current user
func (profileInfo *ProfileInfo) HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserEditInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, err := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser, err := profileInfo.UserApp.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = newUser.UpdateFrom(userInput)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = profileInfo.UserApp.SaveUser(newUser)
	if err != nil {
		switch err.Error() {
		case "Username or email is already taken":
			w.WriteHeader(http.StatusConflict)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteProfile deletes profile of current user, logging them out automatically
func (profileInfo *ProfileInfo) HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userCookie := r.Context().Value("cookieInfo").(*entity.CookieInfo)

	err := profileInfo.CookieApp.RemoveCookie(userCookie)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	err = profileInfo.UserApp.DeleteUser(userCookie.UserID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetProfile returns specified profile
func (profileInfo *ProfileInfo) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	user := new(entity.User)
	var err error
	vars := mux.Vars(r)
	idStr, passedID := vars["id"]
	switch passedID {
	case true:
		{
			id, _ := strconv.Atoi(idStr)
			user, err = profileInfo.UserApp.GetUser(id)
			if err != nil {
				if err.Error() == "No user found with such id" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	case false: // ID was not passed
		{
			username, passedUsername := vars["username"]
			switch passedUsername {
			case true:
				{
					user, err = profileInfo.UserApp.GetUserByUsername(username)
					if err != nil {
						if err.Error() == "No user found with such username" {
							w.WriteHeader(http.StatusNotFound)
							return
						}
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

			case false: // Username was also not passed
				{
					userCookie := r.Context().Value("cookieInfo").(*entity.CookieInfo)
					if userCookie == nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					user, err = profileInfo.UserApp.GetUser(userCookie.UserID)
					if err != nil {
						if err.Error() == "No user found with such id" {
							w.WriteHeader(http.StatusNotFound)
							return
						}
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}

	var userOutput entity.UserOutput
	userOutput.FillFromUser(user)
	userOutput.Password = "" // Password is ommitted on purpose

	responseBody, err := json.Marshal(userOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)
	return
}

var maxPostAvatarBodySize = 8 * 1024 * 1024 // 8 mB
// HandlePostAvatar takes avatar from request and assigns it to current user
func (profileInfo *ProfileInfo) HandlePostAvatar(w http.ResponseWriter, r *http.Request) {
	bodySize := r.ContentLength
	if bodySize < 0 { // No avatar was passed
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if bodySize > int64(maxPostAvatarBodySize) { // Avatar is too large
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(bodySize)
	file, handler, err := r.FormFile("avatarImage")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	newAvatarPath := "avatars/" + handler.Filename // TODO: avatars folder sharding by date, random prefix generato

	err = profileInfo.S3App.UploadFile(file, newAvatarPath)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID
	user, err := profileInfo.UserApp.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Avatar != "/assets/img/default-avatar.jpg" { // TODO: this should be a global variable, probably
		err = profileInfo.S3App.DeleteFile(user.Avatar)

		if err != nil {
			profileInfo.S3App.DeleteFile(newAvatarPath) // deleting newly uploaded avatar
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	user.Avatar = newAvatarPath
	err = profileInfo.UserApp.SaveUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
