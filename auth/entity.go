package auth

import (
	"net/http"
	"sync"
)

type user struct {
	username  string
	password  string // TODO: hashing
	firstName string
	lastName  string
	avatar    string // path to avatar
}

// UserInput if used to parse JSON with users' data
type UserInput struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
}

type usersMap struct {
	users          map[int]user
	lastFreeUserID int
	mu             sync.Mutex
}

type cookieInfo struct {
	userID int
	cookie *http.Cookie
}

type sessionMap struct {
	sessions map[string]cookieInfo // key is cookie value, for easier lookup
	mu       sync.Mutex
}
