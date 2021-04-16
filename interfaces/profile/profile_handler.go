package profile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"pinterest/interfaces/middleware"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ProfileInfo keep information about apps and cookies needed for profile package
type ProfileInfo struct {
	userApp          application.UserAppInterface
	cookieApp        application.CookieAppInterface
	s3App            application.S3AppInterface
	notificationsApp application.NotificationsAppInterface
}

func NewProfileInfo(userApp application.UserAppInterface, cookieApp application.CookieAppInterface,
	s3App application.S3AppInterface, notificationsApp application.NotificationsAppInterface) *ProfileInfo {
	return &ProfileInfo{
		userApp:          userApp,
		cookieApp:        cookieApp,
		s3App:            s3App,
		notificationsApp: notificationsApp,
	}
}

//HandleChangePassword changes password of current user
func (profileInfo *ProfileInfo) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserPassChangeInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, _ := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := profileInfo.userApp.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Password = userInput.Password
	err = profileInfo.userApp.SaveUser(user)
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

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserEditInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, _ := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser, err := profileInfo.userApp.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = newUser.UpdateFrom(userInput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = profileInfo.userApp.SaveUser(newUser)
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

	userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)

	err := profileInfo.cookieApp.RemoveCookie(userCookie)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	err = profileInfo.userApp.DeleteUser(userCookie.UserID)
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
	idStr, passedID := vars[string(entity.IDKey)]
	switch passedID {
	case true:
		{
			id, _ := strconv.Atoi(idStr)
			user, err = profileInfo.userApp.GetUser(id)
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
			username, passedUsername := vars[string(entity.UsernameKey)]
			switch passedUsername {
			case true:
				{
					user, err = profileInfo.userApp.GetUserByUsername(username)
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
					userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)
					if userCookie == nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					user, err = profileInfo.userApp.GetUser(userCookie.UserID)
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

	cookie, found := middleware.CheckCookies(r, profileInfo.cookieApp)
	if !found {
		userOutput.Email = ""
		responseBody, err := json.Marshal(userOutput)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}

	currentUserID := cookie.UserID
	otherUserID := user.UserID
	if currentUserID != otherUserID {
		userOutput.Email = ""
		followed, err := profileInfo.userApp.CheckIfFollowed(currentUserID, otherUserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userOutput.Followed = &followed
	}

	responseBody, err := json.Marshal(userOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
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
	file, _, err := r.FormFile("avatarImage")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = profileInfo.userApp.UpdateAvatar(userID, file)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (profileInfo *ProfileInfo) HandleFollowProfile(w http.ResponseWriter, r *http.Request) {
	var followedUser *entity.User = nil
	var err error // Maybe move this line into switch?
	vars := mux.Vars(r)
	idStr, passedID := vars[string(entity.IDKey)]
	switch passedID {
	case true:
		{
			followedID, _ := strconv.Atoi(idStr)
			followedUser, err = profileInfo.userApp.GetUser(followedID)
			if err != nil {
				if err.Error() == "No user found with such id" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	case false: // ID was not passed
		{
			followedUsername := vars[string(entity.UsernameKey)]
			followedUser, err = profileInfo.userApp.GetUserByUsername(followedUsername)
			if err != nil {
				if err.Error() == "No user found with such id" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	followerID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	followedID := followedUser.UserID
	err = profileInfo.userApp.Follow(followerID, followedID)
	if err != nil {
		if err.Error() == "This follow relation already exists" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	followerUser, err := profileInfo.userApp.GetUser(followerID)
	if err == nil {
		notificationID, err := profileInfo.notificationsApp.AddNotification(&entity.Notification{
			UserID:   followedID,
			Title:    "New follower!",
			Category: "followers",
			Text:     "You have received a new follower: " + followerUser.Username,
			IsRead:   false,
		})
		if err == nil {
			profileInfo.notificationsApp.SendNotification(followedID, notificationID)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (profileInfo *ProfileInfo) HandleUnfollowProfile(w http.ResponseWriter, r *http.Request) {
	var followedUser *entity.User = nil
	var err error
	vars := mux.Vars(r)
	idStr, passedID := vars[string(entity.IDKey)]
	switch passedID {
	case true:
		{
			followedID, _ := strconv.Atoi(idStr)
			followedUser, err = profileInfo.userApp.GetUser(followedID)
			if err != nil {
				if err.Error() == "No user found with such id" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	case false: // ID was not passed
		{
			followedUsername := vars[string(entity.UsernameKey)]
			followedUser, err = profileInfo.userApp.GetUserByUsername(followedUsername)
			if err != nil {
				if err.Error() == "No user found with such id" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	followerID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	followedID := followedUser.UserID
	err = profileInfo.userApp.Unfollow(followerID, followedID)
	if err != nil {
		if err.Error() == "That follow relation does not exist" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	followerUser, err := profileInfo.userApp.GetUser(followerID)
	if err == nil {
		notificationID, err := profileInfo.notificationsApp.AddNotification(&entity.Notification{
			UserID:   followedID,
			Title:    "Follower lost!",
			Category: "followers",
			Text:     "You have  lost a follower: " + followerUser.Username,
			IsRead:   false,
		})
		if err == nil {
			profileInfo.notificationsApp.SendNotification(followedID, notificationID)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
