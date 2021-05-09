package auth

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "google.golang.org/grpc"
	. "pinterest/services/auth/proto"
)

type service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *service {
	return &service{db}
}

func (s *service) LoginUser(ctx context.Context, in *UserAuth) (*CookieInfo, error) {

}

func (s *service) LogoutUser(context.Context, *UserID) (*Error, error) {

}

func (s *service) CheckCookie(context.Context, *Cookie) (*CheckCookieResponse, error) {

}