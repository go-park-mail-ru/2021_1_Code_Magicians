package application

import (
	"context"
	"net/http"
	"pinterest/domain/entity"
	grpcAuth "pinterest/services/auth/proto"
	_ "pinterest/services/user/proto"
	"strings"
)

type AuthApp struct {
	grpcClient grpcAuth.AuthClient
	us         UserAppInterface
	cookieApp  CookieAppInterface
}

func NewAuthApp(grpcClient grpcAuth.AuthClient, us UserAppInterface, cookieApp CookieAppInterface) *AuthApp {
	return &AuthApp{
		grpcClient: grpcClient,
		us:         us,
		cookieApp:  cookieApp,
	}
}

type AuthAppInterface interface {
	CheckUserCredentials(username string, password string) (*entity.CookieInfo, error)
	LogoutUser(userID int) error
	CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) // Check if passed cookie value is present in any active session
}

func (authApp *AuthApp) CheckUserCredentials(username string, password string) (*entity.CookieInfo, error) {
	user, err := authApp.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	_, err = authApp.grpcClient.CheckUserCredentials(context.Background(),
		&grpcAuth.UserAuth{Username: username, Password: password})

	if err != nil { // TODO: hashing
		if strings.Contains(err.Error(), entity.IncorrectPasswordError.Error()) {
			return nil, entity.IncorrectPasswordError
		}
		return nil, err
	}

	cookie, err := authApp.cookieApp.GenerateCookie()
	if err != nil {
		for err == entity.DuplicatingCookieValueError {
			cookie, err = authApp.cookieApp.GenerateCookie()
		}
	}

	resultCookieInfo := &entity.CookieInfo{UserID: user.UserID, Cookie: cookie}
	err = authApp.cookieApp.AddCookieInfo(resultCookieInfo)
	if err != nil {
		return nil, err
	}

	return resultCookieInfo, nil
}

func (authApp *AuthApp) LogoutUser(userID int) error {
	cookieInfo, found := authApp.cookieApp.SearchByUserID(userID)
	if !found {
		return entity.UserNotLoggedInError
	}

	return authApp.cookieApp.RemoveCookie(cookieInfo)
}

func (authApp *AuthApp) CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) {
	grpcCookieInfo, isCookie := authApp.cookieApp.SearchByValue(cookie.Value)

	if !isCookie {
		return nil, false
	}

	return &entity.CookieInfo{
		UserID: int(grpcCookieInfo.UserID),
		Cookie: cookie,
	}, isCookie
}
