package usage

import (
	"log"
	"net/http"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type AuthApp struct {
	us        repository.UserRepository
	cookieApp CookieAppInterface
}

func NewAuthApp(us repository.UserRepository, cookieApp CookieAppInterface) *AuthApp {
	return &AuthApp{
		us:        us,
		cookieApp: cookieApp,
	}
}

type AuthAppInterface interface {
	LoginUser(username string, password string) (*entity.CookieInfo, error)
	LogoutUser(userID int) error
	CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) // Check if passed cookie value is present in any active session
}

func (authApp *AuthApp) LoginUser(username string, password string) (*entity.CookieInfo, error) {
	log.Println("EBELEX1")
	user, err := authApp.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	log.Println("EBELEX2")
	if user.Password != password { // TODO: hashing
		return nil, entity.IncorrectPasswordError
	}

	cookie, err := authApp.cookieApp.GenerateCookie()
	for err == entity.DuplicatingCookieValueError {
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
