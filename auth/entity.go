package auth

import (
	"net/http"
	"sync"
)

// User is, well, a struct depicting a user
type User struct {
	Username  string
	Password  string // TODO: hashing
	FirstName string
	LastName  string
	Email     string
	Avatar    string // path to avatar
}

// UserIO is used to parse JSON with users' data
type UserIO struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
}

// UsersMap is basically a database's fake
type UsersMap struct {
	Users          map[int]User
	LastFreeUserID int
	Mu             sync.Mutex
}

// CookieInfo contains information about a cookie: which user it belongs to and cookie itself
type CookieInfo struct {
	UserID int
	cookie *http.Cookie
}

type sessionMap struct {
	sessions map[string]CookieInfo // key is cookie value, for easier lookup
	mu       sync.Mutex
}
