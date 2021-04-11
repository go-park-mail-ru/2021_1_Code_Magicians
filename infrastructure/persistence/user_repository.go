package persistence

import (
	"context"
	"errors"
	"fmt"
	"log"
	"pinterest/domain/entity"
	"strings"

	"github.com/jackc/pgx/v4"
)

type UserRepo struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *UserRepo {
	return &UserRepo{db}
}

// emptyIfNil replaces nil input with pointer to empty string, noop otherwise
func emptyIfNil(input *string) *string {
	if input == nil {
		return new(string)
	}
	return input
}

const createUserQuery string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, $7)\n" +
	"RETURNING userID"

const createUserQueryDefaulAvatar string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, DEFAULT)\n" +
	"RETURNING userID"

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (r *UserRepo) CreateUser(user *entity.User) (int, error) {
	firstNamePtr := &user.FirstName
	if user.FirstName == "" {
		firstNamePtr = nil
	}
	lastNamePtr := &user.LastName
	if user.LastName == "" {
		lastNamePtr = nil
	}

	var row pgx.Row
	switch user.Avatar {
	case "": // If avatar was not specified, we need to use it's default value
		row = r.db.QueryRow(context.Background(), createUserQueryDefaulAvatar,
			user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr)
	default:
		row = r.db.QueryRow(context.Background(), createUserQuery,
			user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr, user.Avatar)
	}

	newUserID := 0
	err := row.Scan(&newUserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return -1, fmt.Errorf("Username or email is already taken")
		}

		// Other errors
		log.Println(err)
		return -1, err
	}
	return newUserID, nil
}

const saveUserQuery string = "UPDATE Users\n" +
	"SET username=$1, passwordhash=$2, salt=$3, email=$4, first_name=$5, last_name=$6, avatar=$7\n" +
	"WHERE userID=$8"

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (r *UserRepo) SaveUser(user *entity.User) error {
	_, err := r.db.Exec(context.Background(), saveUserQuery, user.Username, user.Password, user.Salt, user.Email,
		user.FirstName, user.LastName, user.Avatar, user.UserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return fmt.Errorf("Username or email is already taken")
		}

		// Other errors
		log.Println(err)
		return err
	}
	return nil
}

const deleteUserQuery string = "DELETE FROM Users WHERE userID=$1"

// SaveUser deletes user with passed ID
// It returns nil on success and error on failure
func (r *UserRepo) DeleteUser(userID int) error {
	commandTag, err := r.db.Exec(context.Background(), deleteUserQuery, userID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("User not found")
	}
	return nil
}

const getUserQuery string = "SELECT username, passwordhash, salt, email, first_name, last_name, avatar FROM Users WHERE userID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *UserRepo) GetUser(userID int) (*entity.User, error) {
	user := entity.User{UserID: userID}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := r.db.QueryRow(context.Background(), getUserQuery, userID)
	err := row.Scan(&user.Username, &user.Password, &user.Salt, &user.Email, &firstNamePtr, &secondNamePtr, &avatarPtr)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No user found with such id")
		}
		// Other errors
		return nil, err
	}

	user.FirstName = *emptyIfNil(firstNamePtr)
	user.LastName = *emptyIfNil(secondNamePtr)
	user.Avatar = *emptyIfNil(avatarPtr)
	return &user, nil
}

const getUsersQuery string = "SELECT userID, username, passwordhash, salt, email, first_name, last_name, avatar FROM Users"

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (r *UserRepo) GetUsers() ([]entity.User, error) {
	users := make([]entity.User, 0)
	rows, err := r.db.Query(context.Background(), getUserQuery)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No users found in database")
		}

		// Other errors
		return nil, err
	}

	for rows.Next() {
		user := entity.User{}
		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err := rows.Scan(&user.UserID, &user.Username, &user.Password, &user.Salt, &user.Email, &firstNamePtr, &secondNamePtr, &avatarPtr)
		if err != nil {
			return nil, err // TODO: error handling
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		users = append(users, user)
	}
	return users, nil
}

const getUserByUsernameQuery string = "SELECT userID, passwordhash, salt, email, first_name, last_name, avatar\n" +
	"FROM Users WHERE username=$1"

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (r *UserRepo) GetUserByUsername(username string) (*entity.User, error) {
	user := entity.User{Username: username}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := r.db.QueryRow(context.Background(), getUserByUsernameQuery, username)
	err := row.Scan(&user.UserID, &user.Password, &user.Salt, &user.Email, &firstNamePtr, &secondNamePtr, &avatarPtr)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No user found with such username")
		}

		// Other errors
		return nil, err
	}

	user.FirstName = *emptyIfNil(firstNamePtr)
	user.LastName = *emptyIfNil(secondNamePtr)
	user.Avatar = *emptyIfNil(avatarPtr)
	return &user, nil
}
