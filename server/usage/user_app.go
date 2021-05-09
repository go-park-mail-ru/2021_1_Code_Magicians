package usage

import (
	"bufio"
	"context"
	"io"
	"log"
	"pinterest/domain/entity"
	grpcUser "pinterest/services/user/proto"
	"time"
)

type UserApp struct {
	grpcClient grpcUser.UserClient
	boardApp   BoardAppInterface
}

func NewUserApp(us grpcUser.UserClient, boardApp BoardAppInterface) *UserApp {
	return &UserApp{us, boardApp}
}

type UserAppInterface interface {
	CreateUser(*entity.User) (int, error) // Create user, returns created user's ID
	SaveUser(*entity.User) error          // Save changed user to database
	ChangePassword(*entity.User) error
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
	userID, err := u.grpcClient.CreateUser(context.Background(), newUser)
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
	newUser := grpcUser.UserEditInput{}
	FillEditForm(user, &newUser)
	_, err := u.grpcClient.SaveUser(context.Background(), &newUser)
	return err
}

func (u *UserApp) ChangePassword(user *entity.User) error {
	_, err := u.grpcClient.ChangePassword(context.Background(),
		&grpcUser.Password{UserID: int64(user.UserID),
			Password: user.Password})
	return err
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (u *UserApp) DeleteUser(userID int) error {
	user, err := u.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		return err
	}

	if user.Avatar != string(entity.AvatarDefaultPath) {
		_, err = u.grpcClient.DeleteFile(context.Background(), &grpcUser.FilePath{ImagePath: user.Avatar})

		if err != nil {
			return err
		}
	}

	_, err = u.grpcClient.DeleteUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
	return err
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUser(userID int) (*entity.User, error) {
	userOutput, err := u.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
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
	usersList, err := u.grpcClient.GetUsers(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	users := ReturnUsersList(usersList.Users)
	return users, nil
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (u *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	userOutput, err := u.grpcClient.GetUserByUsername(context.Background(), &grpcUser.Username{Username: username})
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := u.grpcClient.UpdateAvatar(ctx)
	if err != nil {
		return entity.FileUploadError
	}
	req := &grpcUser.UploadAvatar{
		Data: &grpcUser.UploadAvatar_Extension{
			Extension: extension,
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 8*1024*1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req = &grpcUser.UploadAvatar{
			Data: &grpcUser.UploadAvatar_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	oldAvatarPath := user.Avatar
	user.Avatar = res.Path
	err = u.SaveUser(user)
	if err != nil {
		u.grpcClient.DeleteFile(context.Background(), &grpcUser.FilePath{ImagePath: res.Path})
		return entity.UserSavingError
	}

	if oldAvatarPath != string(entity.AvatarDefaultPath) {
		_, err = u.grpcClient.DeleteFile(ctx, &grpcUser.FilePath{ImagePath: oldAvatarPath})

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
	_, err := u.grpcClient.Follow(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return err
}

func (u *UserApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	_, err := u.grpcClient.Unfollow(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return err
}

func (u *UserApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.SelfFollowError
	}
	isFollowed, err := u.grpcClient.CheckIfFollowed(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return isFollowed.IsFollowed, err
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (u *UserApp) SearchUsers(keyWords string) ([]entity.User, error) {
	usersList, err := u.grpcClient.SearchUsers(context.Background(), &grpcUser.SearchInput{KeyWords: keyWords})
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

func FillEditForm(user *entity.User, userEdit *grpcUser.UserEditInput) {
	userEdit.UserID = int64(user.UserID)
	userEdit.Username = user.Username
	userEdit.Email = user.Email
	userEdit.FirstName = user.FirstName
	userEdit.LastName = user.LastName
	userEdit.Password = user.Password
	userEdit.AvatarLink = user.Avatar
	userEdit.Salt = user.Salt
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

	for _, userOut := range userOutList {
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
