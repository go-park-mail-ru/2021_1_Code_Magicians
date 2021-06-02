package auth

import (
	"context"
	"pinterest/domain/entity"
	. "pinterest/services/auth/proto"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tarantool/go-tarantool"
	_ "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	postgresDB  *pgxpool.Pool
	tarantoolDB *tarantool.Connection
}

func NewService(postgresDB *pgxpool.Pool, tarantoolDB *tarantool.Connection) *service {
	return &service{
		postgresDB:  postgresDB,
		tarantoolDB: tarantoolDB,
	}
}

const GetUserPasswordQuery = "SELECT passwordhash FROM Users WHERE username=$1;"

func (s *service) CheckUserCredentials(ctx context.Context, userCredentials *UserAuth) (*Error, error) {
	tx, err := s.postgresDB.Begin(context.Background())
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
	cookieAsInterface := cookieInfoToInterfaces(cookieInfo)
	resp, err := s.tarantoolDB.Replace("sessions", cookieAsInterface) // If session already exists, we update it
	if err != nil {
		switch resp.Code {
		case tarantool.ErrTupleFound:
			return &Error{}, entity.CookieFoundError
		}
	}

	return &Error{}, nil
}

func (s *service) SearchByValue(ctx context.Context, cookieVal *CookieValue) (*CookieInfo, error) {
	resp, err := s.tarantoolDB.Select("sessions", "secondary", 0, 1, tarantool.IterEq, []interface{}{cookieVal.CookieValue})

	if err != nil {
		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.CookieNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.CookieNotFoundError
	}

	cookieInfo := interfacesToCookieInfo(resp.Tuples()[0])

	if cookieInfo.Cookie.Expires.AsTime().Before(time.Now()) { // We check if cookie is not past it's expiration date
		s.RemoveCookie(ctx, cookieInfo)
		return nil, entity.CookieNotFoundError
	}

	return cookieInfo, nil
}

func (s *service) SearchByUserID(ctx context.Context, userID *UserID) (*CookieInfo, error) {
	resp, err := s.tarantoolDB.Select("sessions", "primary", 0, 1, tarantool.IterEq, []interface{}{userID.Uid})

	if err != nil {
		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.CookieNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.CookieNotFoundError
	}

	cookieInfo := interfacesToCookieInfo(resp.Tuples()[0])

	if cookieInfo.Cookie.Expires.AsTime().Before(time.Now()) { // We check if cookie is not past it's expiration date
		s.RemoveCookie(ctx, cookieInfo)
		return nil, entity.CookieNotFoundError
	}

	return cookieInfo, nil
}

func (s *service) RemoveCookie(ctx context.Context, cookieInfo *CookieInfo) (*Error, error) {
	_, err := s.tarantoolDB.Delete("sessions", "primary", []interface{}{cookieInfo.UserID})
	return &Error{}, err
}

const GetUser = "SELECT passwordhash FROM Users WHERE username=$1;"

func (s *service) CheckUserByVkToken(ctx context.Context, tokenInfo *VkTokenInfo) (*UserID, error) {
	resp, err := s.tarantoolDB.Select("vk_tokens", "primary", 0, 1, tarantool.IterEq, []interface{}{tokenInfo.VkUserID})

	if err != nil {
		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.VkTokenNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.VkTokenNotFoundError
	}

	return &UserID{Uid: int64(resp.Tuples()[0][3].(uint64))}, nil
}

func (s *service) AddVkToken(ctx context.Context, tokenInfo *VkTokenInfo) (*Error, error) {
	tokenInfoAsInterface := vkTokenInfoToInterfaces(tokenInfo)
	resp, err := s.tarantoolDB.Replace("vk_tokens", tokenInfoAsInterface) // If session already exists, we update it
	if err != nil {
		switch resp.Code {
		case tarantool.ErrTupleFound:
			return &Error{}, entity.VkTokenDuplicateError
		}
	}

	return &Error{}, nil
}

func (s *service) RemoveVkToken(ctx context.Context, tokenInfo *VkTokenInfo) (*Error, error) {
	_, err := s.tarantoolDB.Delete("vk_tokens", "primary", []interface{}{tokenInfo.VkUserID})
	return &Error{}, err
}

func cookieInfoToInterfaces(cookieInfo *CookieInfo) []interface{} {
	cookieAsInterfaces := make([]interface{}, 3)
	cookieAsInterfaces[0] = uint(cookieInfo.UserID)
	cookieAsInterfaces[1] = cookieInfo.Cookie.Value
	cookieAsInterfaces[2] = uint(timeToUnixTimestamp(cookieInfo.Cookie.Expires.AsTime()))
	return cookieAsInterfaces
}

func interfacesToCookieInfo(interfaces []interface{}) *CookieInfo {
	cookie := new(Cookie)
	cookie.Value = interfaces[1].(string)
	cookie.Expires = timestamppb.New(unixTimestampToTime(int64(interfaces[2].(uint64))))

	cookieInfo := new(CookieInfo)
	cookieInfo.UserID = int64(interfaces[0].(uint64))
	cookieInfo.Cookie = cookie
	return cookieInfo
}

func vkTokenInfoToInterfaces(tokenInfo *VkTokenInfo) []interface{} {
	tokenInfoAsInterface := make([]interface{}, 4)
	tokenInfoAsInterface[0] = uint(tokenInfo.VkUserID)
	tokenInfoAsInterface[1] = tokenInfo.Token
	tokenInfoAsInterface[2] = uint(timeToUnixTimestamp(tokenInfo.Expires.AsTime()))
	tokenInfoAsInterface[3] = uint(tokenInfo.UserID)
	return tokenInfoAsInterface
}

func unixTimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func timeToUnixTimestamp(timeInput time.Time) int64 {
	return timeInput.Unix()
}
