package application

import (
	"fmt"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type UserApp struct {
	us repository.UserRepository
}

func NewUserApp(us repository.UserRepository) *UserApp {
	return &UserApp{us}
}

type UserAppInterface interface {
	CreateUser(*entity.User) (int, error)             // Create user, returns created user's ID
	SaveUser(*entity.User) error                      // Save changed user to database
	DeleteUser(int) error                             // Delete user with passed userID from database
	GetUser(int) (*entity.User, error)                // Get user by his ID
	GetUsers() ([]entity.User, error)                 // Get all users
	GetUserByUsername(string) (*entity.User, error)   // Get user by his username
	CheckUserCredentials(string, string) (int, error) // Check if passed username and password are correct
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (u *UserApp) CreateUser(user *entity.User) (int, error) {
	return u.us.CreateUser(user)
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (u *UserApp) SaveUser(user *entity.User) error {
	return u.us.SaveUser(user)
}

// SaveUser deletes user with passed ID
// It returns nil on success and error on failure
func (u *UserApp) DeleteUser(userID int) error {
	return u.us.DeleteUser(userID)
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUser(userID int) (*entity.User, error) {
	return u.us.GetUser(userID)
}

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (u *UserApp) GetUsers() ([]entity.User, error) {
	return u.us.GetUsers()
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	return u.us.GetUserByUsername(username)
}

// GetUserCredentials check whether there is user with such username/password pair
// It returns user's ID, nil on success and nil, error on failure
// Those errors are descriptive and tell what did not match
func (u *UserApp) CheckUserCredentials(username string, password string) (int, error) { // TODO: return actual user, will save a request to DB
	user, err := u.us.GetUserByUsername(username)
	if err != nil {
		return -1, err
	}
	if user.Password != password { // TODO: hashing
		return -1, fmt.Errorf("Password does not match")
	}

	return user.UserID, nil
}
