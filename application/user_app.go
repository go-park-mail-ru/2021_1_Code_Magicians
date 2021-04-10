package application

import (
	"fmt"
	"io"
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
	CreateUser(*entity.User) (int, error)                      // Create user, returns created user's ID
	SaveUser(*entity.User) error                               // Save changed user to database
	DeleteUser(int, S3AppInterface) error                      // Delete user with passed userID from database
	GetUser(int) (*entity.User, error)                         // Get user by his ID
	GetUsers() ([]entity.User, error)                          // Get all users
	GetUserByUsername(string) (*entity.User, error)            // Get user by his username
	CheckUserCredentials(string, string) (*entity.User, error) // Check if passed username and password are correct
	UpdateAvatar(int, io.Reader, S3AppInterface) error         // Replace user's avatar with one passed as second parameter
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (u *UserApp) CreateUser(user *entity.User) (int, error) {
	initialBoard := &entity.Board{UserID: user.UserID, Title: "Saved pins"}
	_, err := BoardApp.AddBoard(initialBoard)
	return u.us.CreateUser(user)
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (u *UserApp) SaveUser(user *entity.User) error {
	return u.us.SaveUser(user)
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (u *UserApp) DeleteUser(userID int, s3App S3AppInterface) error {
	user, err := u.us.GetUser(userID)
	if err != nil {
		return err
	}

	if user.Avatar != "/assets/img/default-avatar.jpg" { // TODO: this should be a global variable or s3App's parameter, probably
		err = s3App.DeleteFile(user.Avatar)

		if err != nil {
			return err
		}
	}

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
// It returns user, nil on success and nil, error on failure
// Those errors are descriptive and tell what did not match
func (u *UserApp) CheckUserCredentials(username string, password string) (*entity.User, error) {
	user, err := u.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user.Password != password { // TODO: hashing
		return nil, fmt.Errorf("Password does not match")
	}

	return user, nil
}

func (u *UserApp) UpdateAvatar(userID int, file io.Reader, s3App S3AppInterface) error {
	user, err := u.GetUser(userID)
	if err != nil {
		return fmt.Errorf("Could not find user in database")
	}

	filenamePrefix, err := GenerateRandomString(40) // generating random image
	if err != nil {
		return fmt.Errorf("Could not generate filename")
	}

	newAvatarPath := "avatars/" + filenamePrefix + ".jpg" // TODO: avatars folder sharding by date
	err = s3App.UploadFile(file, newAvatarPath)
	if err != nil {
		return fmt.Errorf("File upload failed")
	}

	oldAvatarPath := user.Avatar
	user.Avatar = newAvatarPath
	err = u.SaveUser(user)
	if err != nil {
		s3App.DeleteFile(newAvatarPath)
		return fmt.Errorf("User saving failed")
	}

	if oldAvatarPath != "/assets/img/default-avatar.jpg" { // TODO: this should be a global variable, probably
		err = s3App.DeleteFile(oldAvatarPath)

		if err != nil {
			s3App.DeleteFile(newAvatarPath) // deleting newly uploaded avatar
			return fmt.Errorf("Old avatar deletion failed")
		}
	}

	return nil
}
