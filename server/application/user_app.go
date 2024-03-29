package application

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"pinterest/domain/entity"
	grpcUser "pinterest/services/user/proto"
	"strconv"
	"strings"
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
	CreateUser(user *entity.User) (int, error)                         // Create user, returns created user's ID
	CreateUserWithVK(tokenInput *entity.UserVkTokenInput) (int, error) // Create user using vk's info
	SaveUser(user *entity.User) error                                  // Save changed user to database
	ChangePassword(user *entity.User) error                            // Change user's password
	DeleteUser(userID int) error                                       // Delete user with passed userID from database
	GetUser(userID int) (*entity.User, error)                          // Get user by his ID
	GetUsers() ([]entity.User, error)                                  // Get all users
	GetUserByUsername(username string) (*entity.User, error)           // Get user by his username
	UpdateAvatar(userID int, file io.Reader, extension string) error   // Replace user's avatar with one passed as second parameter
	SearchUsers(keywords string) ([]entity.User, error)                // Get all users by passed keywords
}

// CreateUser adds new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (userApp *UserApp) CreateUser(user *entity.User) (int, error) {
	newUser := new(grpcUser.UserReg)
	FillRegForm(user, newUser)
	userID, err := userApp.grpcClient.CreateUser(context.Background(), newUser)
	if err != nil {
		if strings.Contains(err.Error(), entity.UsernameEmailDuplicateError.Error()) {
			return -1, entity.UsernameEmailDuplicateError
		}
		return -1, err
	}

	initialBoard := &entity.Board{UserID: int(userID.Uid), Title: "Saved pins", Description: "Fast save"}
	_, err = userApp.boardApp.CreateBoard(initialBoard)
	if err != nil {
		userApp.DeleteUser(user.UserID)
		return -1, err
	}

	return int(userID.Uid), nil
}

// CreateUserWithVK add new user to database with fields from vk's response
// It returns user's assigned ID and nil on success, any number and error on failure
func (userApp *UserApp) CreateUserWithVK(tokenInput *entity.UserVkTokenInput) (int, error) {
	user, err := userApp.parseUserByVkToken(tokenInput)
	if err != nil {
		return -1, err
	}

	return userApp.CreateUser(user)
}

// getUserByToken takes vk token and returns corresponding user using info from vk's servers
func (userApp *UserApp) parseUserByVkToken(tokenInput *entity.UserVkTokenInput) (*entity.User, error) {
	resp, err := http.Get(fmt.Sprintf("%smethod/users.get?params[user_ids]=%d&params[fields]=first_name,last_name&"+
		"params[name_case]=Nom&access_token=%s&v=5.131",
		entity.VkAPIURLKey, tokenInput.VkID, tokenInput.Token))

	if err != nil {
		fmt.Println(err)
		// TODO: error handling
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(data))
		// TODO: proper vk errors handling
		return nil, fmt.Errorf("Could not get token from vk")
	}

	userVkRegInputs := new(entity.UserVkRegInputs)
	err = json.NewDecoder(resp.Body).Decode(userVkRegInputs)
	if err != nil {
		return nil, err
	}
	if len(userVkRegInputs.Users) == 0 {
		return nil, fmt.Errorf("Could not parse user from vk api's response")
	}

	user := new(entity.User)
	user.Username = "vk_user:" + strconv.Itoa(tokenInput.VkID) // TODO: fix username matches with this and usual registrations
	// user.Avatar = userVkFullInput.Avatar
	user.FirstName = userVkRegInputs.Users[0].FirstName
	user.LastName = userVkRegInputs.Users[0].LastName
	user.Email = tokenInput.Email
	user.VkID = tokenInput.VkID
	return user, nil
}

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (userApp *UserApp) SaveUser(user *entity.User) error {
	newUser := grpcUser.UserEditInput{}
	FillEditForm(user, &newUser)
	_, err := userApp.grpcClient.SaveUser(context.Background(), &newUser)
	if err != nil {
		if strings.Contains(err.Error(), entity.UsernameEmailDuplicateError.Error()) {
			return entity.UsernameEmailDuplicateError
		}
		return err
	}

	return nil
}

func (userApp *UserApp) ChangePassword(user *entity.User) error {
	_, err := userApp.grpcClient.ChangePassword(context.Background(),
		&grpcUser.Password{UserID: int64(user.UserID),
			Password: user.Password})
	if err != nil {
		if strings.Contains(err.Error(), entity.UsernameEmailDuplicateError.Error()) {
			return entity.UsernameEmailDuplicateError
		}
		return err
	}

	return nil
}

// SaveUser deletes user with passed ID
// S3AppInterface is needed for avatar deletion
// It returns nil on success and error on failure
func (userApp *UserApp) DeleteUser(userID int) error {
	user, err := userApp.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return entity.UserNotFoundError
		}
		return err
	}

	if user.Avatar != string(entity.UserAvatarDefaultPath) {
		_, err = userApp.grpcClient.DeleteFile(context.Background(), &grpcUser.FilePath{ImagePath: user.Avatar})
		if err != nil {
			return entity.FileDeletionError
		}
	}

	_, err = userApp.grpcClient.DeleteUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return entity.UserNotFoundError
		}
		return err
	}

	return nil
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (userApp *UserApp) GetUser(userID int) (*entity.User, error) {
	userOutput, err := userApp.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(userID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return nil, entity.UserNotFoundError
		}
		return nil, err
	}

	user := new(entity.User)
	FillOutForm(user, userOutput)
	return user, nil
}

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (userApp *UserApp) GetUsers() ([]entity.User, error) {
	usersList, err := userApp.grpcClient.GetUsers(context.Background(), nil)
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return nil, entity.UserNotFoundError
		}
		return nil, err
	}

	users := ReturnUsersList(usersList.Users)
	return users, nil
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
// Those errors are descriptive and tell what did not match
func (userApp *UserApp) GetUserByUsername(username string) (*entity.User, error) {
	userOutput, err := userApp.grpcClient.GetUserByUsername(context.Background(), &grpcUser.Username{Username: username})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return nil, entity.UserNotFoundError
		}
		return nil, err
	}

	user := new(entity.User)
	FillOutForm(user, userOutput)
	return user, nil
}

func (userApp *UserApp) UpdateAvatar(userID int, file io.Reader, extension string) error {
	user, err := userApp.GetUser(userID)
	if err != nil {
		return entity.UserNotFoundError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := userApp.grpcClient.UpdateAvatar(ctx)
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
		return fmt.Errorf("cannot send image info to server: \n%s\n%s", err, stream.RecvMsg(nil))
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 3.5*1024*1024) // jrpc cannot receive packages larger than 4 MB

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read chunk to buffer: \n%s", err)
		}

		req = &grpcUser.UploadAvatar{
			Data: &grpcUser.UploadAvatar_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send chunk to server: \n%s", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("cannot receive response: \n%s", err)
	}

	oldAvatarPath := user.Avatar
	user.Avatar = res.Path
	err = userApp.SaveUser(user)
	if err != nil {
		userApp.grpcClient.DeleteFile(context.Background(), &grpcUser.FilePath{ImagePath: res.Path})

		if strings.Contains(err.Error(), entity.UserSavingError.Error()) {
			return entity.UserSavingError
		}
		return err
	}

	if oldAvatarPath != string(entity.UserAvatarDefaultPath) {
		_, err = userApp.grpcClient.DeleteFile(ctx, &grpcUser.FilePath{ImagePath: oldAvatarPath})
		if err != nil {
			return entity.FileDeletionError
		}
	}

	return nil
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (userApp *UserApp) SearchUsers(keyWords string) ([]entity.User, error) {
	usersList, err := userApp.grpcClient.SearchUsers(context.Background(), &grpcUser.SearchInput{KeyWords: keyWords})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.UsersNotFoundError.Error()):
			return nil, entity.UsersNotFoundError
		case strings.Contains(err.Error(), entity.UserScanError.Error()):
			return nil, entity.UserScanError
		}
		return nil, err
	}

	users := ReturnUsersList(usersList.Users)
	return users, nil
}

func FillRegForm(user *entity.User, userReg *grpcUser.UserReg) {
	userReg.Username = user.Username
	userReg.Email = user.Email
	userReg.FirstName = user.FirstName
	userReg.LastName = user.LastName
	userReg.Password = user.Password
	userReg.VkID = int64(user.VkID)
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
	userEdit.VkID = int64(user.VkID)
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
	user.BoardsCount = int(userOut.BoardsCount)
	user.PinsCount = int(userOut.PinsCount)
	user.VkID = int(userOut.VkID)
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
		user.VkID = int(userOut.VkID)

		userList = append(userList, user)
	}
	return userList
}
