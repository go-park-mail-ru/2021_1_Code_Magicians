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
	Username  *string `json:"username,omitempty"`
	Password  *string `json:"password,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     *string `json:"email,omitempty"`
	Avatar    *string `json:"avatarLink,omitempty"`
}

// FillNillsWithEmptyStr replaces nil pointers in userIO with pointers to empty string
func (userIO *UserIO) FillNilsWithEmptyStr() {
	if userIO.Username == nil {
		userIO.Username = new(string)
	}
	if userIO.Password == nil {
		userIO.Password = new(string)
	}
	if userIO.FirstName == nil {
		userIO.FirstName = new(string)
	}
	if userIO.LastName == nil {
		userIO.LastName = new(string)
	}
	if userIO.Email == nil {
		userIO.Email = new(string)
	}
	if userIO.Avatar == nil {
		userIO.Avatar = new(string)
	}
}

// UpdateUser updates user with values from userIO
func (userIO *UserIO) UpdateUser(user *User) {
	if userIO.Username != nil {
		user.Username = *userIO.Username
	}
	if userIO.Password != nil {
		user.Password = *userIO.Password
	}
	if userIO.FirstName != nil {
		user.FirstName = *userIO.FirstName
	}
	if userIO.LastName != nil {
		user.LastName = *userIO.LastName
	}
	if userIO.Email != nil {
		user.Email = *userIO.Email
	}
	if userIO.Avatar != nil {
		user.Avatar = *userIO.Avatar
	}
}

func (userIO *UserIO) FillFromUser(user *User) {
	if user.Username != "" {
		userIO.Username = &user.Username
	}
	if user.Password != "" {
		userIO.Password = &user.Password
	}
	if user.FirstName != "" {
		userIO.FirstName = &user.FirstName
	}
	if user.LastName != "" {
		userIO.LastName = &user.LastName
	}
	if user.Email != "" {
		userIO.Email = &user.Email
	}
	if user.Avatar != "" {
		userIO.Avatar = &user.Avatar
	}
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
