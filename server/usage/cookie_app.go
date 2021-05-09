package usage

import (
	"net/http"
	"pinterest/domain/entity"
	authProto "pinterest/services/auth/proto"
	"sync"
	"time"
)

type CookieApp struct {
	sessionsByValue  map[string]*entity.CookieInfo
	sessionsByUserID map[int]*entity.CookieInfo // Each value from sessionsByUserID is also sessionsByValue and vice versa
	mu               sync.Mutex
	cookieLength     int
	duration         time.Duration
}

func NewCookieApp(cookieLength int, duration time.Duration) *CookieApp {
	return &CookieApp{
		sessionsByValue:  make(map[string]*entity.CookieInfo),
		sessionsByUserID: make(map[int]*entity.CookieInfo),
		mu:               sync.Mutex{},
		cookieLength:     cookieLength,
		duration:         duration,
	}
}

type CookieAppInterface interface {
	AddCookieInfo(cookieInfo *entity.CookieInfo) error
	SearchByValue(sessionValue string) (*entity.CookieInfo, bool)
	SearchByUserID(userID int) (*entity.CookieInfo, bool)
	RemoveCookie(*entity.CookieInfo) error
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

	cookieApp.mu.Lock()
	cookieApp.sessionsByValue[cookieInfo.Cookie.Value] = &(*cookieInfo) // Copying by value
	cookieApp.sessionsByUserID[cookieInfo.UserID] = cookieApp.sessionsByValue[cookieInfo.Cookie.Value]
	cookieApp.mu.Unlock()
	return nil
}

func (cookieApp *CookieApp) SearchByUserID(userID int) (*entity.CookieInfo, bool) {
	cookieApp.mu.Lock()
	cookieInfo, found := cookieApp.sessionsByUserID[userID]
	cookieApp.mu.Unlock()

	if !found {
		return nil, false
	}

	if cookieInfo.Cookie.Expires.Before(time.Now()) { // We check if cookie is not past it's expiration date
		cookieApp.RemoveCookie(cookieInfo)
		return nil, false
	}

	return cookieInfo, found
}

func (cookieApp *CookieApp) SearchByValue(cookieValue string) (*entity.CookieInfo, bool) {
	cookieApp.mu.Lock()
	cookieInfo, found := cookieApp.sessionsByValue[cookieValue]
	cookieApp.mu.Unlock()

	if !found {
		return nil, false
	}

	if cookieInfo.Cookie.Expires.Before(time.Now()) { // We check if cookie is not past it's expiration date
		cookieApp.RemoveCookie(cookieInfo)
		return nil, false
	}

	return cookieInfo, found
}

func (cookieApp *CookieApp) RemoveCookie(cookieInfo *entity.CookieInfo) error {
	cookieApp.mu.Lock()
	delete(cookieApp.sessionsByValue, cookieInfo.Cookie.Value)
	delete(cookieApp.sessionsByUserID, cookieInfo.UserID)
	cookieApp.mu.Unlock()
	return nil
}

func FillGRPCCookie(cookie *http.Cookie, grpcCookie *authProto.Cookie) {
	cookie.Name = grpcCookie.Name
	cookie.Path = grpcCookie.Path
	cookie.SameSite = http.SameSite(grpcCookie.SameSite)
	cookie.HttpOnly = grpcCookie.HttpOnly
	cookie.Value = grpcCookie.Value
	cookie.Expires = grpcCookie.Expires.AsTime()
	cookie.Secure = grpcCookie.Secure
}
