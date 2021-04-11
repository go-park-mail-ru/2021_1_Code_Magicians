package application

import (
	"crypto/rand"
	"encoding/base64"
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
	GenerateCookie(int, time.Duration) (*http.Cookie, error)
	AddCookie(*entity.CookieInfo) error
	CheckCookie(*http.Cookie) (*entity.CookieInfo, bool)
	RemoveCookie(*entity.CookieInfo) error
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// generateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func (c *CookieApp) GenerateCookie(cookieLength int, duration time.Duration) (*http.Cookie, error) {
	sessionValue, err := GenerateRandomString(cookieLength) // cookie value - random string
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(duration)
	return &http.Cookie{
		Name:     "session_id",
		Value:    sessionValue,
		Path:     "/", // Cookie should be usable on entire website
		Expires:  expirationTime,
		Secure:   true, // We use HTTPS
		HttpOnly: true, // So that frontend won't have direct access to cookies
		SameSite: http.SameSiteNoneMode,
	}, nil
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
