package application

import (
	"context"
	"net/http"
	"os"
	"pinterest/domain/entity"
	grpcAuth "pinterest/services/auth/proto"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CookieApp struct {
	grpcClient   grpcAuth.AuthClient
	cookieLength int
	duration     time.Duration
}

func NewCookieApp(grpcClient grpcAuth.AuthClient, cookieLength int, duration time.Duration) *CookieApp {
	return &CookieApp{
		cookieLength: cookieLength,
		duration:     duration,
		grpcClient:   grpcClient,
	}
}

type CookieAppInterface interface {
	GenerateCookie() (*http.Cookie, error)
	AddCookieInfo(cookieInfo *entity.CookieInfo) error
	SearchByValue(sessionValue string) (*grpcAuth.CookieInfo, bool)
	SearchByUserID(userID int) (*grpcAuth.CookieInfo, bool)
	RemoveCookie(*grpcAuth.CookieInfo) error
}

func (cookieApp *CookieApp) GenerateCookie() (*http.Cookie, error) {
	sessionValue, err := entity.GenerateRandomString(cookieApp.cookieLength) // cookie value - random string
	if err != nil {
		return nil, entity.CookieGenerationError
	}

	expirationTime := time.Now().Add(cookieApp.duration)
	if os.Getenv("HTTPS_ON") == "true" {
		return &http.Cookie{
			Name:     string(entity.CookieNameKey),
			Value:    sessionValue,
			Path:     "/", // Cookie should be usable on entire website
			Expires:  expirationTime,
			Secure:   true, // We use HTTPS
			HttpOnly: true, // So that frontend won't have direct access to cookies
			SameSite: http.SameSiteNoneMode,
		}, nil
	}
	return &http.Cookie{
		Name:     string(entity.CookieNameKey),
		Value:    sessionValue,
		Path:     "/", // Cookie should be usable on entire website
		Expires:  expirationTime,
		HttpOnly: true, // So that frontend won't have direct access to cookies
	}, nil
}

func (cookieApp *CookieApp) SearchByUserID(userID int) (*grpcAuth.CookieInfo, bool) {
	resCookieInfo, err := cookieApp.grpcClient.SearchByUserID(context.Background(), &grpcAuth.UserID{Uid: int64(userID)})
	if err != nil {
		return nil, false
	}

	return resCookieInfo, true
}

func (cookieApp *CookieApp) SearchByValue(cookieValue string) (*grpcAuth.CookieInfo, bool) {
	resCookieInfo, err := cookieApp.grpcClient.SearchByValue(context.Background(), &grpcAuth.CookieValue{CookieValue: cookieValue})
	if err != nil {
		return nil, false
	}

	return resCookieInfo, true
}

func (cookieApp *CookieApp) AddCookieInfo(cookieInfo *entity.CookieInfo) error {
	oldCookieInfo, found := cookieApp.SearchByUserID(cookieInfo.UserID)
	if found {
		cookieApp.RemoveCookie(oldCookieInfo)
	}

	_, found = cookieApp.SearchByValue(cookieInfo.Cookie.Value)
	if found {
		return entity.DuplicatingCookieValueError
	}

	grpcCookie := grpcAuth.Cookie{}
	FillGRPCCookie(&grpcCookie, cookieInfo.Cookie)
	grpcCookieInfo := grpcAuth.CookieInfo{UserID: int64(cookieInfo.UserID), Cookie: &grpcCookie}
	_, err := cookieApp.grpcClient.AddCookieInfo(context.Background(), &grpcCookieInfo)

	return err // Will actually almost always be nil
}

func (cookieApp *CookieApp) RemoveCookie(cookieInfo *grpcAuth.CookieInfo) error {
	_, err := cookieApp.grpcClient.RemoveCookie(context.Background(), cookieInfo)

	return err
}

func FillGRPCCookie(grpcCookie *grpcAuth.Cookie, cookie *http.Cookie) {
	grpcCookie.Value = cookie.Value
	grpcCookie.Expires = timestamppb.New(cookie.Expires)
}
