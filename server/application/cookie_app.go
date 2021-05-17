package application

import (
	"context"
	"net/http"
	"os"
	"pinterest/domain/entity"
	authProto "pinterest/services/auth/proto"
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
	SearchByValue(sessionValue string) (*authProto.CookieInfo, bool)
	SearchByUserID(userID int) (*authProto.CookieInfo, bool)
	RemoveCookie(*authProto.CookieInfo) error
}

func (cookieApp *CookieApp) GenerateCookie() (*http.Cookie, error) {
	sessionValue, err := entity.GenerateRandomString(cookieApp.cookieLength) // cookie value - random string
	if err != nil {
		return nil, entity.CookieGenerationError
	}

	expirationTime := time.Now().Add(cookieApp.duration)
	if os.Getenv("HTTPS_ON") == "true" {
		return &http.Cookie{
			Name:     entity.CookieNameKey,
			Value:    sessionValue,
			Path:     "/", // Cookie should be usable on entire website
			Expires:  expirationTime,
			Secure:   true, // We use HTTPS
			HttpOnly: true, // So that frontend won't have direct access to cookies
			SameSite: http.SameSiteNoneMode,
		}, nil
	}
	return &http.Cookie{
		Name:     entity.CookieNameKey,
		Value:    sessionValue,
		Path:     "/", // Cookie should be usable on entire website
		Expires:  expirationTime,
		HttpOnly: true, // So that frontend won't have direct access to cookies
	}, nil
}

func (cookieApp *CookieApp) SearchByUserID(userID int) (*authProto.CookieInfo, bool) {
	resCookieInfo, _ := cookieApp.grpcClient.SearchByUserID(context.Background(), &grpcAuth.UserID{Uid: int64(userID)})

	return resCookieInfo.CookieInfo, resCookieInfo.IsCookie
}

func (cookieApp *CookieApp) SearchByValue(cookieValue string) (*authProto.CookieInfo, bool) {
	resCookieInfo, _ := cookieApp.grpcClient.SearchByValue(context.Background(), &grpcAuth.CookieValue{CookieValue: cookieValue})

	return resCookieInfo.CookieInfo, resCookieInfo.IsCookie
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

	grpcCookie := authProto.Cookie{}
	FillGRPCCookie(&grpcCookie, cookieInfo.Cookie)
	grpcCookieInfo := authProto.CookieInfo{UserID: int64(cookieInfo.UserID), Cookie: &grpcCookie}
	_, err := cookieApp.grpcClient.AddCookieInfo(context.Background(), &grpcCookieInfo)
	return err
}

func (cookieApp *CookieApp) RemoveCookie(cookieInfo *authProto.CookieInfo) error {
	_, err := cookieApp.grpcClient.RemoveCookie(context.Background(), cookieInfo)
	return err
}

func FillGRPCCookie(grpcCookie *authProto.Cookie, cookie *http.Cookie) {
	grpcCookie.Value = cookie.Value
	grpcCookie.Expires = timestamppb.New(cookie.Expires)
}
