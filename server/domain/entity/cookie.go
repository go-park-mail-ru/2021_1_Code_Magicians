package entity

import (
	"net/http"
	"sync"
)

// CookieInfo contains information about a cookie: which user it belongs to and cookie itself
type CookieInfo struct {
	UserID int
	Cookie *http.Cookie
}

// SessionMap is used to keep track of users currently logged in
type SessionMap struct {
	Sessions map[string]CookieInfo // key is cookie value, for easier lookup
	Mu       sync.Mutex
}
