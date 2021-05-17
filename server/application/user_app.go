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
	CreateUser(user *entity.User) (int, error)                       // Create user, returns created user's ID
	SaveUser(user *entity.User) error                                // Save changed user to database
	DeleteUser(userID int) error                                     // Delete user with passed userID from database
	GetUser(userID int) (*entity.User, error)                        // Get user by his ID
	GetUsers() ([]entity.User, error)                                // Get all users
	GetUserByUsername(username string) (*entity.User, error)         // Get user by his username
	UpdateAvatar(userID int, file io.Reader, extension string) error // Replace user's avatar with one passed as second parameter
	Follow(followerID int, followedID int) error                     // Make first user follow second
	Unfollow(followerID int, followedID int) error                   // Make first user unfollow second
	CheckIfFollowed(followerID int, followedID int) (bool, error)    // Check if first user follows second. Err != nil if those users are the same
	SearchUsers(keywords string) ([]entity.User, error)              // Get all users by passed keywords
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (userApp *UserApp) CreateUser(user *entity.User) (int, error) {
	userID, err := userApp.us.CreateUser(user)
	if err != nil {
		return -1, err
	}

	initialBoard := &entity.Board{UserID: userID, Title: "Saved pins", Description: "Fast save"}
	_, err = userApp.boardApp.AddBoard(initialBoard)
	if err != nil {

		_ = userApp.DeleteUser(user.UserID)
		return -1, err
	}

	return userID, nil
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (userApp *UserApp) SaveUser(user *entity.User) error {
	return userApp.us.SaveUser(user)
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (userApp *UserApp) DeleteUser(userID int) error {
	user, err := userApp.us.GetUser(userID)
	if err != nil {
		return err
	}

	if user.Avatar != string(entity.AvatarDefaultPath) {
		err = userApp.s3App.DeleteFile(user.Avatar)

		if err != nil {
			return err
		}
	}

	return userApp.us.DeleteUser(userID)
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (userApp *UserApp) GetUser(userID int) (*entity.User, error) {
	return userApp.us.GetUser(userID)
}

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (userApp *UserApp) GetUsers() ([]entity.User, error) {
	return userApp.us.GetUsers()
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (userApp *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	return userApp.us.GetUserByUsername(username)
}

// GetUserCredentials check whether there is user with such username/password pair
// It returns user, nil on success and nil, error on failure
// Those errors are descriptive and tell what did not match
func (userApp *UserApp) CheckUserCredentials(username string, password string) (*entity.User, error) {
	user, err := userApp.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user.Password != password { // TODO: hashing
		return nil, entity.IncorrectPasswordError
	}

	return user, nil
}

func (userApp *UserApp) UpdateAvatar(userID int, file io.Reader, extension string) error {
	user, err := userApp.GetUser(userID)
	if err != nil {
		return entity.UserNotFoundError
	}

	filenamePrefix, err := GenerateRandomString(40) // generating random filename
	if err != nil {
		return entity.FilenameGenerationError
	}

	newAvatarPath := "avatars/" + filenamePrefix + extension // TODO: avatars folder sharding by date
	err = userApp.s3App.UploadFile(file, newAvatarPath)
	if err != nil {
		return entity.FileUploadError
	}

	oldAvatarPath := user.Avatar
	user.Avatar = newAvatarPath
	err = userApp.SaveUser(user)
	if err != nil {
		userApp.s3App.DeleteFile(newAvatarPath)
		return entity.UserSavingError
	}

	if oldAvatarPath != string(entity.AvatarDefaultPath) {
		err = userApp.s3App.DeleteFile(oldAvatarPath)

		if err != nil {
			return entity.FileDeletionError
		}
	}

	return nil
}

func (userApp *UserApp) Follow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return userApp.us.Follow(followerID, followedID)
}

func (userApp *UserApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return userApp.us.Unfollow(followerID, followedID)
}

func (userApp *UserApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.SelfFollowError
	}
	return userApp.us.CheckIfFollowed(followerID, followedID)
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (userApp *UserApp) SearchUsers(keyWords string) ([]entity.User, error) {
	return userApp.us.SearchUsers(keyWords)
}
