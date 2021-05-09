package usage

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
	LoginUser(username string, password string) (*entity.CookieInfo, error)
	LogoutUser(userID int) error
	CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) // Check if passed cookie value is present in any active session
}

func (authApp *AuthApp) LoginUser(username string, password string) (*entity.CookieInfo, error) {
	user, err := authApp.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	_, err = authApp.grpcClient.LoginUser(context.Background(),
		&grpcAuth.UserAuth{Username: username, Password: password})
	if err != nil { // TODO: hashing
		return nil, err
	}

	cookie, err := authApp.cookieApp.GenerateCookie()
	for strings.Contains(err.Error(), entity.DuplicatingCookieValueError.Error()) {
		cookie, err = authApp.cookieApp.GenerateCookie()
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
	return authApp.cookieApp.SearchByValue(cookie.Value)
}
