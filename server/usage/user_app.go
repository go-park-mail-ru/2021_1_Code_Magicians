package usage

import (
	"io"
	"log"
	"pinterest/domain/entity"
	grpcUser "pinterest/services/user/proto"
)

type UserApp struct {
	grpcClient grpcUser.UserClient
	boardApp   BoardAppInterface
	s3App      S3AppInterface
}

func NewUserApp(us grpcUser.UserClient, boardApp BoardAppInterface, s3App S3AppInterface) *UserApp {
	return &UserApp{us, boardApp, s3App}
}

type UserAppInterface interface {
	CreateUser(*entity.User) (int, error)           // Create user, returns created user's ID
	SaveUser(*entity.User) error                    // Save changed user to database
	DeleteUser(int) error                           // Delete user with passed userID from database
	GetUser(int) (*entity.User, error)              // Get user by his ID
	GetUsers() ([]entity.User, error)               // Get all users
	GetUserByUsername(string) (*entity.User, error) // Get user by his username
	UpdateAvatar(int, io.Reader, string) error      // Replace user's avatar with one passed as second parameter
	Follow(int, int) error                          // Make first user follow second
	Unfollow(int, int) error                        // Make first user unfollow second
	CheckIfFollowed(int, int) (bool, error)         // Check if first user follows second. Err != nil if those users are the same
	SearchUsers(string) ([]entity.User, error)      // Get all users by passed keywords
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (u *UserApp) CreateUser(user *entity.User) (int, error) {
	newUser := new(grpcUser.UserReg)
	FillRegForm(user, newUser)
	userID, err := u.grpcClient.CreateUser(nil, newUser)
	if err != nil {
		return -1, err
	}

	initialBoard := &entity.Board{UserID: int(userID.Uid), Title: "Saved pins", Description: "Fast save"}
	_, err = u.boardApp.AddBoard(initialBoard)
	if err != nil {

		_ = u.DeleteUser(user.UserID)
		return -1, err
	}

	return int(userID.Uid), nil
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (u *UserApp) SaveUser(user *entity.User) error {
	newUser := new(grpcUser.UserReg)
	FillRegForm(user, newUser)
	_, err := u.grpcClient.SaveUser(nil, newUser)
	return err
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (u *UserApp) DeleteUser(userID int) error {
	user, err := u.grpcClient.GetUser(nil, &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		return err
	}

	if user.Avatar != string(entity.AvatarDefaultPath) {
		err = u.s3App.DeleteFile(user.Avatar)

		if err != nil {
			return err
		}
	}

	_, err = u.grpcClient.DeleteUser(nil, &grpcUser.UserID{Uid: int64(userID)})
	return err
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUser(userID int) (*entity.User, error) {
	log.Println("EEEEEEEEE")
	userOutput, err := u.grpcClient.GetUser(nil, &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		return nil, err
	}
	user := new(entity.User)
	FillOutForm(user, userOutput)
	return user, err
}

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (u *UserApp) GetUsers() ([]entity.User, error) {
	usersList, err := u.grpcClient.GetUsers(nil, nil)
	if err != nil {
		return nil, err
	}
	users := ReturnUsersList(usersList.Users)
	return users, nil
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	userOutput, err := u.grpcClient.GetUserByUsername(nil, &grpcUser.Username{Username: username})
	if err != nil {
		return nil, err
	}

	user := new(entity.User)
	FillOutForm(user, userOutput)
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
	_, err := u.grpcClient.Follow(nil, &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return err
}

func (u *UserApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	_, err := u.grpcClient.Unfollow(nil, &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return err
}

func (u *UserApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.SelfFollowError
	}
	isFollowed, err := u.grpcClient.CheckIfFollowed(nil, &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return isFollowed.IsFollowed, err
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (u *UserApp) SearchUsers(keyWords string) ([]entity.User, error) {
	usersList, err := u.grpcClient.SearchUsers(nil, &grpcUser.SearchInput{KeyWords: keyWords})
	users := ReturnUsersList(usersList.Users)
	return users, err
}

func FillRegForm(user *entity.User, userReg *grpcUser.UserReg) {
	userReg.Username = user.Username
	userReg.Email = user.Email
	userReg.FirstName = user.FirstName
	userReg.LastName = user.LastName
	userReg.Password = user.Password
}

func FillOutForm(user *entity.User, userOut *grpcUser.UserOutput) {
	user.UserID = int(userOut.UserID)
	user.Username = userOut.Username
	user.Email = userOut.Email
	user.FirstName = userOut.FirstName
	user.LastName = userOut.LastName
	user.Avatar = userOut.Avatar
	user.Following = int(userOut.Following)
	user.FollowedBy = int(userOut.FollowedBy)
}


func ReturnUsersList(userOutList []*grpcUser.UserOutput) []entity.User {
	userList := make([]entity.User, 0)

	for _, userOut:= range userOutList {
		user := entity.User{}
		user.UserID = int(userOut.UserID)
		user.Username = userOut.Username
		user.Email = userOut.Email
		user.FirstName = userOut.FirstName
		user.LastName = userOut.LastName
		user.Avatar = userOut.Avatar
		user.Following = int(userOut.Following)
		user.FollowedBy = int(userOut.FollowedBy)

		userList = append(userList, user)
	}
	return userList
}