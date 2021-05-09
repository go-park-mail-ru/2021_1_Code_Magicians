package auth

import (
	"context"
	"crypto/rand"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"os"
	"pinterest/domain/entity"
	. "pinterest/services/auth/proto"
	"time"
)

type service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *service {
	return &service{db}
}

const GetUserPasswordQuery = "SELECT passwordhash FROM Users WHERE username=$1;"

func (s *service) LoginUser(ctx context.Context, userCredentials *UserAuth) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	password := new(string)

	row := tx.QueryRow(context.Background(), GetUserPasswordQuery, userCredentials.Username)
	err = row.Scan(password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Error{}, entity.UserNotFoundError
		}

		return &Error{},  err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	if *password != userCredentials.Password {
		return &Error{}, entity.IncorrectPasswordError
	}
	return &Error{}, nil
}

func (s *service) LogoutUser(ctx context.Context, userID *UserID) (*Error, error) {
	return nil, nil
}

func (s *service) CheckCookie(ctx context.Context, cookie *Cookie) (*CheckCookieResponse, error) {
	return nil, nil
}

func (s *service) GenerateCookie(ctx context.Context, nothing *empty.Empty) (*Cookie, error) {
	sessionValue, err := entity.GenerateRandomString(40) // cookie value - random string
	if err != nil {
		return nil, entity.CookieGenerationError
	}

	expirationTime := time.Now().Add(10*time.Hour)
	if os.Getenv("HTTPS_ON") == "true" {
		return &Cookie{
			Name:     entity.CookieNameKey,
			Value:    sessionValue,
			Path:     "/", // Cookie should be usable on entire website
			Expires:  timestamppb.New(expirationTime),
			Secure:   true, // We use HTTPS
			HttpOnly: true, // So that frontend won't have direct access to cookies
			SameSite: int64(http.SameSiteNoneMode),
		}, nil
	}
	return &Cookie{
		Name:     entity.CookieNameKey,
		Value:    sessionValue,
		Path:     "/", // Cookie should be usable on entire website
		Expires:  timestamppb.New(expirationTime),
		HttpOnly: true, // So that frontend won't have direct access to cookies
	}, nil
}

func (s *service) AddCookieInfo(ctx context.Context, cookieInfo *CookieInfo) (*Error, error) {
	return nil, nil
}
func (s *service) SearchByValue(ctx context.Context, cookie *Cookie) (*CheckCookieResponse, error) {
	return nil, nil
}
func (s *service) SearchByUserID(ctx context.Context, userID *UserID) (*CheckCookieResponse, error) {
	return nil, nil
}
func (s *service) RemoveCookie(ctx context.Context, cookieInfo *CookieInfo) (*Error, error) {
	return nil, nil
}


// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

