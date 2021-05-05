package application

import (
	"io"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type UserApp struct {
	us       repository.UserRepository
	boardApp BoardAppInterface
	s3App    S3AppInterface
}

func NewUserApp(us repository.UserRepository, boardApp BoardAppInterface, s3App S3AppInterface) *UserApp {
	return &UserApp{us, boardApp, s3App}
}

type UserAppInterface interface {
	CreateUser(*entity.User) (int, error)                      // Create user, returns created user's ID
	SaveUser(*entity.User) error                               // Save changed user to database
	DeleteUser(int) error                                      // Delete user with passed userID from database
	GetUser(int) (*entity.User, error)                         // Get user by his ID
	GetUsers() ([]entity.User, error)                          // Get all users
	GetUserByUsername(string) (*entity.User, error)            // Get user by his username
	CheckUserCredentials(string, string) (*entity.User, error) // Check if passed username and password are correct
	UpdateAvatar(int, io.Reader, string) error                 // Replace user's avatar with one passed as second parameter
	Follow(int, int) error                                     // Make first user follow second
	Unfollow(int, int) error                                   // Make first user unfollow second
	CheckIfFollowed(int, int) (bool, error)                    // Check if first user follows second. Err != nil if those users are the same
	SearchUsers(string) ([]entity.User, error)                 // Get all users by passed keywords
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (u *UserApp) CreateUser(user *entity.User) (int, error) {
	userID, err := u.us.CreateUser(user)
	if err != nil {
		return -1, err
	}

	initialBoard := &entity.Board{UserID: userID, Title: "Saved pins", Description: "Fast save"}
	_, err = u.boardApp.AddBoard(initialBoard)
	if err != nil {

		_ = u.DeleteUser(user.UserID)
		return -1, err
	}

	return userID, nil
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (u *UserApp) SaveUser(user *entity.User) error {
	return u.us.SaveUser(user)
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (u *UserApp) DeleteUser(userID int) error {
	user, err := u.us.GetUser(userID)
	if err != nil {
		return err
	}

	if user.Avatar != string(entity.AvatarDefaultPath) {
		err = u.s3App.DeleteFile(user.Avatar)

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
		return nil, entity.IncorrectPasswordError
	}

	return user, nil
}

func (u *UserApp) UpdateAvatar(userID int, file io.Reader, extension string) error {
	user, err := u.GetUser(userID)
	if err != nil {
		return entity.UserNotFoundError
	}

	filenamePrefix, err := GenerateRandomString(40) // generating random filename
	if err != nil {
		return entity.FilenameGenerationError
	}

	newAvatarPath := "avatars/" + filenamePrefix + extension // TODO: avatars folder sharding by date
	err = u.s3App.UploadFile(file, newAvatarPath)
	if err != nil {
		return entity.FileUploadError
	}

	oldAvatarPath := user.Avatar
	user.Avatar = newAvatarPath
	err = u.SaveUser(user)
	if err != nil {
		u.s3App.DeleteFile(newAvatarPath)
		return entity.UserSavingError
	}

	if oldAvatarPath != string(entity.AvatarDefaultPath) {
		err = u.s3App.DeleteFile(oldAvatarPath)

		if err != nil {
			return entity.FileDeletionError
		}
	}

	return nil
}

func (u *UserApp) Follow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return u.us.Follow(followerID, followedID)
}

func (u *UserApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return u.us.Unfollow(followerID, followedID)
}

func (u *UserApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.FollowThemselfError
	}
	return u.us.CheckIfFollowed(followerID, followedID)
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (u *UserApp) SearchUsers(keyWords string) ([]entity.User, error) {
	return u.us.SearchUsers(keyWords)
}