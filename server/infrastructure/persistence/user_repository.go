package persistence

import (
	"context"
	"log"
	"pinterest/domain/entity"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
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
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
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
	switch user.Avatar {
	case "": // If avatar was not specified, we need to use it's default value
		row = tx.QueryRow(context.Background(), createUserQueryDefaulAvatar,
			user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr)
	default:
		row = tx.QueryRow(context.Background(), createUserQuery,
			user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr, user.Avatar)
	}

	newUserID := 0
	err = row.Scan(&newUserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return -1, entity.UsernameEmailDuplicateError
		}

		// Other errors
		log.Println(err)
		return -1, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return newUserID, nil
}

const saveUserQuery string = "UPDATE Users\n" +
	"SET username=$1, passwordhash=$2, salt=$3, email=$4, first_name=$5, last_name=$6, avatar=$7\n" +
	"WHERE userID=$8"

// SaveUser saves user to database with passed fields
// It returns nil on success and error on failure
func (r *UserRepo) SaveUser(user *entity.User) error {
	tx, err := r.db.Begin(context.Background())
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

const deleteUserQuery string = "DELETE FROM Users WHERE userID=$1"

// SaveUser deletes user with passed ID
// It returns nil on success and error on failure
func (r *UserRepo) DeleteUser(userID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deleteUserQuery, userID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.UserNotFoundError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const getUserQuery string = "SELECT username, passwordhash, salt, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users WHERE userID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *UserRepo) GetUser(userID int) (*entity.User, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	user := entity.User{UserID: userID}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := tx.QueryRow(context.Background(), getUserQuery, userID)
	err = row.Scan(&user.Username, &user.Password, &user.Salt, &user.Email, &firstNamePtr,
		&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}
		// Other errors
		return nil, err
	}

	user.FirstName = *emptyIfNil(firstNamePtr)
	user.LastName = *emptyIfNil(secondNamePtr)
	user.Avatar = *emptyIfNil(avatarPtr)

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &user, nil
}

const getUsersQuery string = "SELECT userID, username, passwordhash, salt, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users"

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (r *UserRepo) GetUsers() ([]entity.User, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	users := make([]entity.User, 0)
	rows, err := tx.Query(context.Background(), getUsersQuery)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}

		// Other errors
		return nil, err
	}

	for rows.Next() {
		user := entity.User{}
		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err := rows.Scan(&user.UserID, &user.Username, &user.Password, &user.Salt, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
		if err != nil {
			return nil, err // TODO: error handling
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		users = append(users, user)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return users, nil
}

const getUserByUsernameQuery string = "SELECT userID, passwordhash, salt, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users WHERE username=$1"

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (r *UserRepo) GetUserByUsername(username string) (*entity.User, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	user := entity.User{Username: username}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := tx.QueryRow(context.Background(), getUserByUsernameQuery, username)
	err = row.Scan(&user.UserID, &user.Password, &user.Salt, &user.Email, &firstNamePtr,
		&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}

		// Other errors
		return nil, err
	}

	user.FirstName = *emptyIfNil(firstNamePtr)
	user.LastName = *emptyIfNil(secondNamePtr)
	user.Avatar = *emptyIfNil(avatarPtr)

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &user, nil
}

const followQuery string = "INSERT INTO Followers(followerID, followedID) VALUES ($1, $2)"
const updateFollowingQuery string = "UPDATE Users SET following = following + 1 WHERE userID=$1"
const updateFollowedByQuery string = "UPDATE Users SET followed_by = followed_by + 1 WHERE userID=$1"

func (r *UserRepo) Follow(followerID int, followedID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	_, err = tx.Exec(context.Background(), followQuery, followerID, followedID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return entity.FollowAlreadyExistsError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_followed"`) {
			return entity.UserNotFoundError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_follower"`) { // Actually does not usually happen because of checks in middleware
			return entity.UserNotFoundError
		}

		return err
	}

	_, err = tx.Exec(context.Background(), updateFollowingQuery, followerID)
	if err != nil {
		log.Println(err)
		return entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateFollowedByQuery, followedID)
	if err != nil {
		return entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const unfollowQuery string = "DELETE FROM Followers WHERE followerID=$1 AND followedID=$2"
const updateUnfollowingQuery string = "UPDATE Users SET following = following - 1 WHERE userID=$1"
const updateUnfollowedByQuery string = "UPDATE Users SET followed_by = followed_by - 1 WHERE userID=$1"

func (r *UserRepo) Unfollow(followerID int, followedID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	result, _ := tx.Exec(context.Background(), unfollowQuery, followerID, followedID)

	if result.RowsAffected() != 1 {
		return entity.FollowNotFoundError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowingQuery, followerID)
	if err != nil {
		return entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowedByQuery, followedID)
	if err != nil {
		return entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return err
}

const checkIfFollowedQuery string = "SELECT 1 FROM Followers WHERE followerID=$1 AND followedID=$2" // returns 1 if found, no rows otherwise

func (r *UserRepo) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return false, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), checkIfFollowedQuery, followerID, followedID)

	var resultingOne int
	err = row.Scan(&resultingOne)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		// Other errors
		return false, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return false, entity.TransactionCommitError
	}
	return true, nil
}
