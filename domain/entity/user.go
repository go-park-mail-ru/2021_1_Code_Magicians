package entity

import (
	"errors"
	"net/http"
	"regexp"
	"sync"

	"github.com/asaskevich/govalidator"
)

const usernameRegexp = "^[a-zA-Z][a-zA-Z0-9_]{1,41}$"
const firstNameRegexp = "^[a-zA-Z ]{0,42}$"

// init initiates custom validators for User struct
func init() {
	govalidator.CustomTypeTagMap.Set("filepath", func(i interface{}, context interface{}) bool {
		matched := false
		switch i.(type) {
		case string:
			matched = govalidator.IsUnixFilePath(i.(string))
		}

		return matched
	})

	govalidator.CustomTypeTagMap.Set("name", func(i interface{}, context interface{}) bool {
		matched := false
		switch i.(type) {
		case string:
			matched, _ = regexp.MatchString(firstNameRegexp, i.(string))
		}

		return matched
	})

	govalidator.CustomTypeTagMap.Set("username", func(i interface{}, context interface{}) bool {
		matched := false
		switch i.(type) {
		case string:
			matched, _ = regexp.MatchString(usernameRegexp, i.(string))
		}

		return matched
	})
}

// User is, well, a struct depicting a user
type User struct {
	UserID    int
	Username  string
	Password  string // TODO: hashing
	FirstName string
	LastName  string
	Email     string
	Avatar    string // path to avatar
	Salt      string
}

// UserOutput is used to marshal JSON with users' data
type UserOutput struct {
	UserID    int    `json:"ID"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Avatar    string `json:"avatarLink,omitempty"`
}

// UserRegInput is used when parsing JSON in auth/signup handler
type UserRegInput struct {
	Username  string `json:"username" valid:"username"`
	Password  string `json:"password" valid:"stringlength(8|30)"`
	Email     string `json:"email" valid:"email"`
	FirstName string `json:"firstName" valid:"name,optional"`
	LastName  string `json:"lastName" valid:"name,optional"`
}

// UserLoginInput is used when parsing JSON in auth/login handler
type UserLoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserPassChangeInput is used when parsing JSON in profile/password handler
type UserPassChangeInput struct {
	Password string `json:"password" valid:"stringlength(8|30)"`
}

// UserEditInput is used when parsing JSON in profile/edit handler
type UserEditInput struct {
	Username  string `json:"username" valid:"username,optional"`
	Email     string `json:"email" valid:"email,optional"`
	FirstName string `json:"firstName" valid:"name,optional"`
	LastName  string `json:"lastName" valid:"name,optional"`
	Avatar    string `json:"avatarLink" valid:"filepath,optional"`
}

// Validate validates UserRegInput struct according to following rules:
// Username - 2-42 alphanumeric, "_" or " " characters
// Password - 8-30 characters
// Email - standard email validity check
// Username uniqueness is NOT checked
func (userInput *UserRegInput) Validate() (bool, error) {
	return govalidator.ValidateStruct(*userInput)
}

// Validate validates UserPassChangeInput struct - Password is 8-30 characters
func (userInput *UserPassChangeInput) Validate() (bool, error) {
	return govalidator.ValidateStruct(*userInput)
}

// Validate validates UserEditInput struct according to following rules:
// Username - 2-42 alphanumeric, "_" or whitespace characters
// LastName, FirstName - 0-42 alpha or whitespace characters
// Email - standard email validity check
// Avatar - some Unix file path
// Username uniqueness or Avatar actual existance are NOT checked
func (userInput *UserEditInput) Validate() (bool, error) {
	return govalidator.ValidateStruct(*userInput)
}

// UpdateFrom changes user fields with non-empty fields of userInput
// By default it's assumed that userInput is validated
func (user *User) UpdateFrom(userInput interface{}) error {
	switch userInput.(type) {
	case *UserRegInput:
		{
			userRegInput := *userInput.(*UserRegInput)
			user.Username = userRegInput.Username
			user.Password = userRegInput.Password // TODO: hashing
			user.Email = userRegInput.Email
			user.FirstName = userRegInput.FirstName
			user.LastName = userRegInput.LastName
		}
	case *UserPassChangeInput:
		user.Password = userInput.(*UserPassChangeInput).Password // TODO: hashing
	case *UserEditInput:
		{
			userEditInput := *userInput.(*UserEditInput)
			if userEditInput.Username != "" {
				user.Username = userEditInput.Username
			}
			if userEditInput.FirstName != "" {
				user.FirstName = userEditInput.FirstName
			}
			if userEditInput.LastName != "" {
				user.LastName = userEditInput.LastName
			}
			if userEditInput.Email != "" {
				user.Email = userEditInput.Email
			}
			if userEditInput.Avatar != "" {
				user.Avatar = userEditInput.Avatar
			}
		}
	default:
		return errors.New("auth.UpdateFrom: Unknown input type")
	}

	return nil
}

func (userOutput *UserOutput) FillFromUser(user *User) {
	userOutput.UserID = user.UserID
	userOutput.Username = user.Username
	userOutput.FirstName = user.FirstName
	userOutput.LastName = user.LastName
	userOutput.Avatar = user.Avatar
}

// // UsersMap is basically a database's fake
// type UsersMap struct {
// 	Users          map[int]User
// 	LastFreeUserID int
// 	Mu             sync.Mutex
// }

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
