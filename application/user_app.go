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
	CreateUser(*entity.User) (int, error) // Returns created user's ID
	SaveUser(*entity.User) error
	DeleteUser(int) error
	GetUser(int) (*entity.User, error) // Get user by his ID
	GetUsers() ([]entity.User, error)  // Get all users
	GetUserByUsername(string) (*entity.User, error)
	CheckUserCredentials(string, string) (int, error) // Check if passed username and password are correct
}

func (u *UserApp) CreateUser(user *entity.User) (int, error) {
	return u.us.CreateUser(user)
}

func (u *UserApp) SaveUser(user *entity.User) error {
	return u.us.SaveUser(user)
}

func (u *UserApp) DeleteUser(userID int) error {
	return u.us.DeleteUser(userID)
}

func (u *UserApp) GetUser(userID int) (*entity.User, error) {
	return u.us.GetUser(userID)
}

func (u *UserApp) GetUsers() ([]entity.User, error) {
	return u.us.GetUsers()
}

func (u *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	return u.us.GetUserByUsername(username)
}

func (u *UserApp) CheckUserCredentials(username string, password string) (int, error) {
	user, err := u.us.GetUserByUsername(username)
	if err != nil {
		return -1, err
	}
	if user.Password != password { // TODO: hashing
		return -1, fmt.Errorf("Password does not match")
	}

	return user.UserID, nil
}
