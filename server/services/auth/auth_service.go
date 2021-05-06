package auth

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc"
	"log"
	"pinterest/domain/entity"
	. "pinterest/services/auth/proto"
	"strings"
)

type service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *service {
	return &service{db}
}


const createUserQuery string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, $7)\n" +
	"RETURNING userID"

const createUserQueryDefaulAvatar string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, DEFAULT)\n" +
	"RETURNING userID"

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (s *service) CreateUser(ctx context.Context, us *UserReg) (*UserID, error) {
	user := entity.User{
		UserID:     0,
		Username:   us.Username,
		Password:   us.Password,
		FirstName:  us.FirstName,
		LastName:   us.LastName,
		Email:      us.Email,
		Avatar:     "",
		Salt:       "", // TODO salt realize
		Following:  0,
		FollowedBy: 0,
	}
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	firstNamePtr := &user.FirstName
	if user.FirstName == "" {
		firstNamePtr = nil
	}
	lastNamePtr := &user.LastName
	if user.LastName == "" {
		lastNamePtr = nil
	}

	var row pgx.Row
	row = tx.QueryRow(context.Background(), createUserQueryDefaulAvatar,
		user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr)

	newUserID := 0
	err = row.Scan(&newUserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return nil, entity.UsernameEmailDuplicateError
		}

		// Other errors
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}

	return &UserID{Uid: int32(newUserID)}, nil
}

func (s *service) SaveUser(ctx context.Context, in *UserReg) (*Error, error) {
	user := entity.User{
		UserID:     0,
		Username:   us.Username,
		Password:   us.Password,
		FirstName:  us.FirstName,
		LastName:   us.LastName,
		Email:      us.Email,
		Avatar:     "",
		Salt:       "", // TODO salt realize
		Following:  0,
		FollowedBy: 0,
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), saveUserQuery, user.Username, user.Password, user.Salt, user.Email,
		user.FirstName, user.LastName, user.Avatar, user.UserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return entity.UsernameEmailDuplicateError
		}

		// Other errors
		log.Println(err)
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

func (user *entity.User) FillFromRegForm(us *UserReg)  {


}