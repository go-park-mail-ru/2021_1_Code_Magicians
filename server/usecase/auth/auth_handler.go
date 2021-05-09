package auth

import (
	"encoding/json"
	"net/http"
	"pinterest/delivery"
	"pinterest/domain/entity"
	"pinterest/usecase/middleware"
	"time"

	"go.uber.org/zap"
)

// AuthInfo keep information about apps and cookies needed for auth package
type AuthInfo struct {
	userApp      delivery.UserAppInterface
	authApp      delivery.AuthAppInterface
	s3App        delivery.S3AppInterface
	boardApp     delivery.BoardAppInterface     // For initial user's board
	websocketApp delivery.WebsocketAppInterface // For setting CSRF token during  login
	logger       *zap.Logger
}

func NewAuthInfo(userApp delivery.UserAppInterface, authApp delivery.AuthAppInterface,
	s3App delivery.S3AppInterface, boardApp delivery.BoardAppInterface,
	websocketApp delivery.WebsocketAppInterface, logger *zap.Logger) *AuthInfo {
	return &AuthInfo{
		userApp:      userApp,
		authApp:      authApp,
		s3App:        s3App,
		boardApp:     boardApp,
		websocketApp: websocketApp,
		logger:       logger,
	}
}

// HandleCreateUser creates user with parameters passed in JSON
func (info *AuthInfo) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	userInput := new(entity.UserRegInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, _ := userInput.Validate()
	if !valid {
		info.logger.Info(entity.ValidationError.Error(),
			zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newUser entity.User
	err = newUser.UpdateFrom(userInput)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newUser.UserID, err = info.userApp.CreateUser(&newUser)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		if err == entity.UsernameEmailDuplicateError {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookieInfo, err := info.authApp.LoginUser(newUser.Username, newUser.Password)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		info.userApp.DeleteUser(newUser.UserID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Replacing token in websocket connection info
	token := r.Header.Get("X-CSRF-Token")
	err = info.websocketApp.ChangeToken(newUser.UserID, token)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookieInfo.Cookie)
	w.WriteHeader(http.StatusCreated)
}

// HandleLoginUser logs user in using provided username and password
func (info *AuthInfo) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	userInput := new(entity.UserLoginInput)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookieInfo, err := info.authApp.LoginUser(userInput.Username, userInput.Password)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.IncorrectPasswordError:
			w.WriteHeader(http.StatusUnauthorized)
		case entity.UserNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Replacing token in websocket connection info
	token := r.Header.Get("X-CSRF-Token")
	err = info.websocketApp.ChangeToken(cookieInfo.UserID, token)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", cookieInfo.UserID),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookieInfo.Cookie)
	w.WriteHeader(http.StatusNoContent)
}

// HandleLogoutUser logs current user out of their session
func (info *AuthInfo) HandleLogoutUser(w http.ResponseWriter, r *http.Request) {
	userCookie := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo)

	err := info.authApp.LogoutUser(userCookie.UserID)
	if err != nil {
		info.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userCookie.UserID),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCookie.Cookie.Expires = time.Now().AddDate(0, 0, -1) // Making cookie expire
	http.SetCookie(w, userCookie.Cookie)

	w.WriteHeader(http.StatusNoContent)
}

// HandleCheckUser checks if current user is logged in
func (info *AuthInfo) HandleCheckUser(w http.ResponseWriter, r *http.Request) {
	_, found := middleware.CheckCookies(r, info.authApp)
	if !found {
		info.logger.Info(entity.UnauthorizedError.Error(),
			zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
