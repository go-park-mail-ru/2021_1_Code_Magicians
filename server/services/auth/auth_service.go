package auth

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "google.golang.org/grpc"
	"pinterest/domain/entity"
	. "pinterest/services/auth/proto"
	"sync"
	"time"
)

type service struct {
	db               *pgxpool.Pool
	sessionsByValue  map[string]*CookieInfo
	sessionsByUserID map[int64]*CookieInfo // Each value from sessionsByUserID is also sessionsByValue and vice versa
	mu               sync.Mutex
}

func NewService(db *pgxpool.Pool) *service {
	return &service{
		db:               db,
		sessionsByValue:  make(map[string]*CookieInfo),
		sessionsByUserID: make(map[int64]*CookieInfo),
		mu:               sync.Mutex{},
	}
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

		return &Error{}, err
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

func (s *service) AddCookieInfo(ctx context.Context, cookieInfo *CookieInfo) (*Error, error) {
	s.mu.Lock()
	s.sessionsByValue[cookieInfo.Cookie.Value] = &(*cookieInfo) // Copying by value
	s.sessionsByUserID[cookieInfo.UserID] = s.sessionsByValue[cookieInfo.Cookie.Value]
	s.mu.Unlock()

	return &Error{}, nil
}

func (s *service) SearchByValue(ctx context.Context, cookieVal *CookieValue) (*CheckCookieResponse, error) {
	s.mu.Lock()
	cookieInfo, found := s.sessionsByValue[cookieVal.CookieValue]
	s.mu.Unlock()

	if !found {
		return &CheckCookieResponse{CookieInfo: nil, IsCookie: false}, nil
	}

	if cookieInfo.Cookie.Expires.AsTime().Before(time.Now()) { // We check if cookie is not past it's expiration date
		s.RemoveCookie(ctx, cookieInfo)
		return &CheckCookieResponse{CookieInfo: nil, IsCookie: false}, nil
	}

	return &CheckCookieResponse{CookieInfo: cookieInfo, IsCookie: found}, nil
}

func (s *service) SearchByUserID(ctx context.Context, userID *UserID) (*CheckCookieResponse, error) {
	s.mu.Lock()
	cookieInfo, found := s.sessionsByUserID[userID.Uid]
	s.mu.Unlock()

	if !found {
		return &CheckCookieResponse{CookieInfo: nil, IsCookie: false}, nil
	}

	if cookieInfo.Cookie.Expires.AsTime().Before(time.Now()) { // We check if cookie is not past it's expiration date
		s.RemoveCookie(ctx, cookieInfo)
		return &CheckCookieResponse{CookieInfo: nil, IsCookie: false}, nil
	}
	return &CheckCookieResponse{CookieInfo: cookieInfo, IsCookie: found}, nil
}

func (s *service) RemoveCookie(ctx context.Context, cookieInfo *CookieInfo) (*Error, error) {
	s.mu.Lock()
	delete(s.sessionsByValue, cookieInfo.Cookie.Value)
	delete(s.sessionsByUserID, cookieInfo.UserID)
	s.mu.Unlock()
	return &Error{}, nil
}
