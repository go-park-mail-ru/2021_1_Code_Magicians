package follow

import (
	"encoding/json"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type FollowInfo struct {
	userApp         application.UserAppInterface
	followApp       application.FollowAppInterface
	notificationApp application.NotificationAppInterface
	logger          *zap.Logger
}

func NewFollowInfo(userApp application.UserAppInterface, followApp application.FollowAppInterface,
	notificationApp application.NotificationAppInterface,
	logger *zap.Logger) *FollowInfo {
	return &FollowInfo{
		userApp:         userApp,
		followApp:       followApp,
		notificationApp: notificationApp,
		logger:          logger,
	}
}

func (followInfo *FollowInfo) HandleFollowProfile(w http.ResponseWriter, r *http.Request) {
	var followedUser *entity.User = nil
	var err error // Maybe move this line into switch?
	vars := mux.Vars(r)
	idStr, passedID := vars[string(entity.IDKey)]
	followerID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	switch passedID {
	case true:
		{
			followedID, _ := strconv.Atoi(idStr)
			followedUser, err = followInfo.userApp.GetUser(followedID)
			if err != nil {
				followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
					zap.Int("for user", followerID), zap.String("method", r.Method))
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
			followedUsername := vars[string(entity.UsernameKey)]
			followedUser, err = followInfo.userApp.GetUserByUsername(followedUsername)
			if err != nil {
				followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
					zap.Int("for user", followerID), zap.String("method", r.Method))
				if err == entity.UserNotFoundError {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	followedID := followedUser.UserID
	err = followInfo.followApp.Follow(followerID, followedID)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", followerID), zap.String("method", r.Method))
		if err == entity.FollowAlreadyExistsError {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	followerUser, err := followInfo.userApp.GetUser(followerID)
	if err == nil {
		notificationID, err := followInfo.notificationApp.AddNotification(&entity.Notification{
			UserID:   followedID,
			Title:    "New follower!",
			Category: "followers",
			Text:     "You have received a new follower: " + followerUser.Username,
			IsRead:   false,
		})
		if err == nil {
			followInfo.notificationApp.SendNotification(followedID, notificationID) // It's alright if notification could not be sent
		} else {
			followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
				zap.Int("for user", followerID), zap.String("method", r.Method))
		}
	} else {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", followerID), zap.String("method", r.Method))
	}

	w.WriteHeader(http.StatusNoContent)
}

func (followInfo *FollowInfo) HandleUnfollowProfile(w http.ResponseWriter, r *http.Request) {
	var followedUser *entity.User = nil
	var err error
	vars := mux.Vars(r)
	idStr, passedID := vars[string(entity.IDKey)]
	followerID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	switch passedID {
	case true:
		{
			followedID, _ := strconv.Atoi(idStr)
			followedUser, err = followInfo.userApp.GetUser(followedID)
			if err != nil {
				followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
					zap.Int("for user", followerID), zap.String("method", r.Method))
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
			followedUsername := vars[string(entity.UsernameKey)]
			followedUser, err = followInfo.userApp.GetUserByUsername(followedUsername)
			if err != nil {
				followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
					zap.Int("for user", followerID), zap.String("method", r.Method))
				if err == entity.UserNotFoundError {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	followedID := followedUser.UserID
	err = followInfo.followApp.Unfollow(followerID, followedID)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", followerID), zap.String("method", r.Method))
		if err == entity.FollowNotFoundError {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	followerUser, err := followInfo.userApp.GetUser(followerID)
	if err == nil {
		notificationID, err := followInfo.notificationApp.AddNotification(&entity.Notification{
			UserID:   followedID,
			Title:    "Follower lost!",
			Category: "followers",
			Text:     "You have  lost a follower: " + followerUser.Username,
			IsRead:   false,
		})
		if err == nil {
			followInfo.notificationApp.SendNotification(followedID, notificationID)
		} else {
			followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
				zap.Int("for user", followerID), zap.String("method", r.Method))
		}
	} else {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", followerID), zap.String("method", r.Method))
	}

	w.WriteHeader(http.StatusNoContent)
}

func (followInfo *FollowInfo) HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, _ := vars[string(entity.IDKey)]
	id, _ := strconv.Atoi(idStr)

	followers, err := followInfo.followApp.GetAllFollowers(id)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.UserNotFoundError, entity.UsersNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	usersOutput := new(entity.UserListOutput)

	for _, user := range followers {
		var userOutput entity.UserOutput
		userOutput.FillFromUser(&user)
		userOutput.Email = "" // Emails are private and should not be passed to unrelated users
		usersOutput.Users = append(usersOutput.Users, userOutput)
	}

	responseBody, err := json.Marshal(usersOutput)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "usage/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (followInfo *FollowInfo) HandleGetFollowed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, _ := vars[string(entity.IDKey)]
	id, _ := strconv.Atoi(idStr)

	followedUsers, err := followInfo.followApp.GetAllFollowed(id)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.UserNotFoundError, entity.UsersNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	usersOutput := new(entity.UserListOutput)

	for _, user := range followedUsers {
		var userOutput entity.UserOutput
		userOutput.FillFromUser(&user)
		userOutput.Email = "" // Emails are private and should not be passed to unrelated users
		usersOutput.Users = append(usersOutput.Users, userOutput)
	}

	responseBody, err := json.Marshal(usersOutput)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "usage/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (followInfo *FollowInfo) HandleGetFollowedPinsList(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	resultPins, err := followInfo.followApp.GetPinsOfFollowedUsers(userID)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(resultPins) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range resultPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	responseBody, err := json.Marshal(pins)
	if err != nil {
		followInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "usage/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
