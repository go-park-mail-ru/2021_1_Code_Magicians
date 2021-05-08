package usage

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	grpcUser "pinterest/services/user/proto"
)

type UserUsage struct {
	grpcClient grpcUser.UserClient
	boardApp   BoardAppInterface
	s3App      S3AppInterface
}

func NewUserUsage(us grpcUser.UserClient, boardApp BoardAppInterface, s3App S3AppInterface) *UserApp {
	return &UserApp{us, boardApp, s3App}
}

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (s *UserUsage) CreateUser(ctx context.Context, us *grpcUser.UserReg) (*grpcUser.UserID, error) {
	return nil, nil
}

func (s *UserUsage) SaveUser(ctx context.Context, us *grpcUser.UserReg) (*grpcUser.Error, error) {
	return nil, nil
}

// DeleteUser deletes user with passed ID
// It returns nil on success and error on failure
func (s *UserUsage) DeleteUser(ctx context.Context, userID *grpcUser.UserID) (*grpcUser.Error, error) {
	return nil, nil
}

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (s *UserUsage) GetUser(ctx context.Context, userID *grpcUser.UserID) (*grpcUser.UserOutput, error) {
	return nil, nil
}

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (s *UserUsage) GetUsers(ctx context.Context, in *empty.Empty) (*grpcUser.UsersListOutput, error) {
	return nil, nil
}

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (s *UserUsage) GetUserByUsername(ctx context.Context, username *grpcUser.Username) (*grpcUser.UserOutput, error) {
	return nil, nil
}

func (s *UserUsage) Follow(ctx context.Context, follows *grpcUser.Follows) (*grpcUser.Error, error) {

	return nil, nil
}

func (s *UserUsage) Unfollow(ctx context.Context, follows *grpcUser.Follows) (*grpcUser.Error, error) {
	return nil, nil
}

func (s *UserUsage) CheckIfFollowed(ctx context.Context, follows *grpcUser.Follows) (*grpcUser.IfFollowedResponse, error) {
	return nil, nil
}

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (s *UserUsage) SearchUsers(ctx context.Context, keyWords *grpcUser.SearchInput) (*grpcUser.UsersListOutput, error) {
	return nil, nil
}

func (s *UserUsage) UpdateAvatar(ctx context.Context) (grpcUser.User_UpdateAvatarServer, error){
	return nil, nil
}
