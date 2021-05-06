package auth

import (
	"context"
	"google.golang.org/grpc"
	"pinterest/domain/entity"

	"github.com/golang/protobuf/ptypes/empty"

)

type AuthClient interface {
	CreateUser(ctx context.Context, userInput entity.UserRegInput) (int, error)
	SaveUser(ctx context.Context, userInput entity.UserRegInput) error
	DeleteUser(ctx context.Context, userID int) error
	GetUser(ctx context.Context, userID int) (*entity.UserOutput, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.UserOutput, error)
	GetUsers(ctx context.Context) (entity.UserListOutput, error)
	CheckUserCredentials(ctx context.Context, userAuth entity.UserLoginInput) (*entity.UserOutput, error)
	UpdateAvatar(ctx context.Context) (Auth_UpdateAvatarClient, error)
	Follow(ctx context.Context, in *Follows, opts ...grpc.CallOption) (*Error, error)
	Unfollow(ctx context.Context, in *Follows, opts ...grpc.CallOption) (*Error, error)
	CheckIfFollowed(ctx context.Context, in *Follows, opts ...grpc.CallOption) (*IfFollowedResponse, error)
	GenerateCookie(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Cookie, error)
	AddCookie(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error)
	CheckCookie(ctx context.Context, in *Cookie, opts ...grpc.CallOption) (*CheckCookieResponse, error)
	RemoveCookie(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error)
}
type service struct {
	authService protoAuth.AuthClient
}

func NewService(authService protoAuth.AuthClient) ServiceAuth {
	return &service{
		authService: authService,
	}
}
