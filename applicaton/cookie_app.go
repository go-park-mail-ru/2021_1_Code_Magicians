package application

import (
	"net/http"
	"pinterest/domain/entity"
	"sync"
	"time"
)

type CookieApp struct {
	sessions map[string]entity.CookieInfo
	mu       sync.Mutex
}

func NewCookieApp() *CookieApp {
	return &CookieApp{sessions: make(map[string]entity.CookieInfo)}
}

type CookieAppInterface interface {
	AddCookie(*entity.CookieInfo) error
	CheckCookie(*http.Cookie) (*entity.CookieInfo, bool)
	RemoveCookie(*entity.CookieInfo) error
}

func (c *CookieApp) AddCookie(cookieInfo *entity.CookieInfo) error {
	c.mu.Lock()
	c.sessions[cookieInfo.Cookie.Value] = *cookieInfo
	c.mu.Unlock()
	return nil
}

func (c *CookieApp) CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) {
	c.mu.Lock()
	userCookieInfo, ok := c.sessions[cookie.Value]
	c.mu.Unlock()

	if !ok { // cookie was not found
		return nil, false
	}

	if userCookieInfo.Cookie.Expires.Before(time.Now()) { // We check our cookie because client could change their expiration date
		c.RemoveCookie(&userCookieInfo)
		return nil, false
	}

	return &userCookieInfo, true
}

func (c *CookieApp) RemoveCookie(cookieInfo *entity.CookieInfo) error {
	c.mu.Lock()
	delete(c.sessions, cookieInfo.Cookie.Value)
	c.mu.Unlock()

	return nil
}
