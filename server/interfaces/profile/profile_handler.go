package profile

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"pinterest/application"
	"pinterest/domain/entity"
	"pinterest/interfaces/middleware"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

// ProfileInfo keep information about apps and cookies needed for profile package
type ProfileInfo struct {
	userApp         application.UserAppInterface
	authApp         application.AuthAppInterface
	cookieApp       application.CookieAppInterface
	followApp       application.FollowAppInterface
	s3App           application.S3AppInterface
	notificationApp application.NotificationAppInterface
	logger          *zap.Logger
}

func NewProfileInfo(userApp application.UserAppInterface, authApp application.AuthAppInterface, cookieApp application.CookieAppInterface,
	followApp application.FollowAppInterface, s3App application.S3AppInterface, notificationApp application.NotificationAppInterface,
	logger *zap.Logger) *ProfileInfo {
	return &ProfileInfo{
		userApp:         userApp,
		authApp:         authApp,
		cookieApp:       cookieApp,
		followApp:       followApp,
		s3App:           s3App,
		notificationApp: notificationApp,
		logger:          logger,
	}
}

//HandleChangePassword changes password of current user
func (profileInfo *ProfileInfo) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserPassChangeInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, _ := userInput.Validate()
	if !valid {
		profileInfo.logger.Info(entity.ValidationError.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := profileInfo.userApp.GetUser(userID)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Password = userInput.Password
	err = profileInfo.userApp.ChangePassword(user)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleEditProfile edits profile of current user
func (profileInfo *ProfileInfo) HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(entity.UserEditInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, _ := userInput.Validate()
	if !valid {
		profileInfo.logger.Info(entity.ValidationError.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser, err := profileInfo.userApp.GetUser(userID)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = newUser.UpdateFrom(userInput)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = profileInfo.userApp.SaveUser(newUser)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		switch err {
		case entity.UsernameEmailDuplicateError:
			w.WriteHeader(http.StatusConflict)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteProfile deletes profile of current user, logging them out automatically
func (profileInfo *ProfileInfo) HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)

	err := profileInfo.authApp.LogoutUser(userCookie.UserID)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userCookie.UserID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	err = profileInfo.userApp.DeleteUser(userCookie.UserID)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userCookie.UserID), zap.String("method", r.Method))
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
				profileInfo.logger.Info(err.Error(),
					zap.String("url", r.RequestURI),
					zap.String("method", r.Method))
				if err == entity.UserNotFoundError {
					w.WriteHeader(http.StatusNotFound)
					return
				}
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
						profileInfo.logger.Info(err.Error(),
							zap.String("url", r.RequestURI),
							zap.String("method", r.Method))
						if err == entity.UserNotFoundError {
							w.WriteHeader(http.StatusNotFound)
							return
						}
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

			case false: // Username was also not passed
				{
					userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)
					if userCookie == nil {
						profileInfo.logger.Info(entity.GetCookieFromContextError.Error(),
							zap.String("url", r.RequestURI),
							zap.String("method", r.Method))
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					user, err = profileInfo.userApp.GetUser(userCookie.UserID)
					if err != nil {
						profileInfo.logger.Info(err.Error(),
							zap.String("url", r.RequestURI),
							zap.String("method", r.Method))
						if err == entity.UserNotFoundError {
							w.WriteHeader(http.StatusNotFound)
							return
						}
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}

	var userOutput entity.UserOutput
	userOutput.FillFromUser(user)

	cookie, found := middleware.CheckCookies(r, profileInfo.authApp)
	if !found {
		userOutput.Email = ""
		responseBody, err := json.Marshal(userOutput)
		if err != nil {
			profileInfo.logger.Info(err.Error(),
				zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
		return
	}

	currentUserID := cookie.UserID
	otherUserID := user.UserID
	if currentUserID != otherUserID {
		userOutput.Email = ""
		followed, err := profileInfo.followApp.CheckIfFollowed(currentUserID, otherUserID)
		if err != nil {
			profileInfo.logger.Info(err.Error(),
				zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userOutput.Followed = &followed
	}

	responseBody, err := json.Marshal(userOutput)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
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

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	if bodySize <= 0 { // No avatar was passed
		profileInfo.logger.Info(entity.NoPicturePassed.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if bodySize > int64(maxPostAvatarBodySize) { // Avatar is too large
		profileInfo.logger.Info(entity.TooLargePicture.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(bodySize)
	file, header, err := r.FormFile("avatarImage")
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	extension := filepath.Ext(header.Filename)
	err = profileInfo.userApp.UpdateAvatar(userID, file, extension)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (profileInfo *ProfileInfo) HandleGetProfilesByKeyWords(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)[string(entity.SearchKeyQuery)]

	keyString = strings.NewReplacer("+", " ").Replace(keyString)
	users, err := profileInfo.userApp.SearchUsers(strings.ToLower(keyString))
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(users) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	usersOutput := new(entity.UserListOutput)

	for _, user := range users {
		var userOutput entity.UserOutput
		userOutput.FillFromUser(&user)
		usersOutput.Users = append(usersOutput.Users, userOutput)
	}

	responseBody, err := json.Marshal(usersOutput)
	if err != nil {
		profileInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
