package user

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_"google.golang.org/grpc"
	_ "google.golang.org/grpc"
	"log"
	"pinterest/domain/entity"
	. "pinterest/services/user/proto"
	"strings"
)

type service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *service {
	return &service{db}
}

const createUserQueryDefaulAvatar string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, DEFAULT)\n" +
	"RETURNING userID"

// CreateUser add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (s *service) CreateUser(ctx context.Context, us *UserReg) (*UserID, error) {
	user := FillFromRegForm(us)

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

	return &UserID{Uid: int64(newUserID)}, nil
}

const saveUserQuery string = "UPDATE Users\n" +
	"SET username=$1, passwordhash=$2, salt=$3, email=$4, first_name=$5, last_name=$6, avatar=$7\n" +
	"WHERE userID=$8"

func (s *service) SaveUser(ctx context.Context, us *UserReg) (*Error, error) {
	user := FillFromRegForm(us)

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), saveUserQuery, user.Username, user.Password, user.Salt, user.Email,
		user.FirstName, user.LastName, user.Avatar, user.UserID)
	if err != nil {
		// If username/email is already taken
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return nil, entity.UsernameEmailDuplicateError
		}

		// Other errors
		log.Println(err)
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return nil, nil
}

const deleteUserQuery string = "DELETE FROM Users WHERE userID=$1"

// DeleteUser deletes user with passed ID
// It returns nil on success and error on failure
func (s *service) DeleteUser(ctx context.Context, userID *UserID) (*Error, error)  {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deleteUserQuery, int(userID.Uid))
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, entity.UserNotFoundError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return nil, nil
}

const getUserQuery string = "SELECT username, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users WHERE userID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (s* service) GetUser(ctx context.Context, userID *UserID) (*UserOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	user := UserOutput{UserID: userID.Uid}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := tx.QueryRow(context.Background(), getUserQuery, userID)
	err = row.Scan(&user.Username, &user.Email, &firstNamePtr,
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

const getUsersQuery string = "SELECT userID, username, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users"

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (s* service) GetUsers(ctx context.Context, in *empty.Empty) (*UsersListOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	users := make([]*UserOutput, 0)
	rows, err := tx.Query(context.Background(), getUsersQuery)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}

		return nil, err
	}

	for rows.Next() {
		user := UserOutput{}
		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
		if err != nil {
			return nil, err // TODO: error handling
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		users = append(users, &user)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &UsersListOutput{Users: users}, nil
}

const getUserByUsernameQuery string = "SELECT userID, email, first_name, last_name, avatar, followed_by, following\n" +
	"FROM Users WHERE username=$1"

// GetUserByUsername fetches user with passed username from database
// It returns that user, nil on success and nil, error on failure
func (s *service) GetUserByUsername(ctx context.Context, username *Username) (*UserOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	user := UserOutput{Username: username.Username}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := tx.QueryRow(context.Background(), getUserByUsernameQuery, username)
	err = row.Scan(&user.UserID, &user.Email, &firstNamePtr,
		&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}

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

func (s *service) Follow(ctx context.Context, follows *Follows) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	_, err = tx.Exec(context.Background(), followQuery, follows.FollowerID, follows.FollowedID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return nil, entity.FollowAlreadyExistsError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_followed"`) {
			return nil, entity.UserNotFoundError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_follower"`) { // Actually does not usually happen because of checks in middleware
			return nil, entity.UserNotFoundError
		}

		return nil, err
	}

	_, err = tx.Exec(context.Background(), updateFollowingQuery, follows.FollowerID)
	if err != nil {
		log.Println(err)
		return nil, entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateFollowedByQuery, follows.FollowedID)
	if err != nil {
		return nil, entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return nil, nil
}

const unfollowQuery string = "DELETE FROM Followers WHERE followerID=$1 AND followedID=$2"
const updateUnfollowingQuery string = "UPDATE Users SET following = following - 1 WHERE userID=$1"
const updateUnfollowedByQuery string = "UPDATE Users SET followed_by = followed_by - 1 WHERE userID=$1"

func (s *service) Unfollow(ctx context.Context, follows *Follows) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	result, _ := tx.Exec(context.Background(), unfollowQuery, follows.FollowerID, follows.FollowedID)

	if result.RowsAffected() != 1 {
		return nil, entity.FollowNotFoundError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowingQuery, follows.FollowerID)
	if err != nil {
		return nil, entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowedByQuery, follows.FollowedID)
	if err != nil {
		return nil, entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return nil, err
}

const checkIfFollowedQuery string = "SELECT 1 FROM Followers WHERE followerID=$1 AND followedID=$2" // returns 1 if found, no rows otherwise

func (s *service) CheckIfFollowed(ctx context.Context, follows *Follows) (*IfFollowedResponse, error) {

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &IfFollowedResponse{IsFollowed: false}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), checkIfFollowedQuery, follows.FollowerID, follows.FollowedID)

	var resultingOne int
	err = row.Scan(&resultingOne)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &IfFollowedResponse{IsFollowed: false}, nil
		}
		// Other errors
		return &IfFollowedResponse{IsFollowed: false}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &IfFollowedResponse{IsFollowed: false}, entity.TransactionCommitError
	}
	return &IfFollowedResponse{IsFollowed: true}, nil
}

const SearchUsersQuery string = "SELECT userID, username, passwordhash, salt, email, first_name, last_name, avatar, followed_by, following FROM Users\n" +
	"WHERE LOWER(username) LIKE $1;"

// SearchUsers fetches all users from database suitable with passed keywords
// It returns slice of users and nil on success, nil and error on failure
func (s *service) SearchUsers(ctx context.Context, keyWords *SearchInput) (*UsersListOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	users := make([]*UserOutput, 0)
	rows, err := tx.Query(context.Background(), SearchUsersQuery,"%" + keyWords.KeyWords + "%")
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.NoResultSearch
		}
		return nil, err
	}

	for rows.Next() {
		user := UserOutput{}

		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following)
		if err != nil {
			return nil, entity.SearchingError
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		users = append(users, &user)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &UsersListOutput{Users: users}, nil
}

func (s *service) UpdateAvatar(Auth_UpdateAvatarServer) error {
	return nil
}

func FillFromRegForm(us *UserReg) entity.User  {
	return entity.User{
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
}

// emptyIfNil replaces nil input with pointer to empty string, noop otherwise
func emptyIfNil(input *string) *string {
	if input == nil {
		return new(string)
	}
	return input
}



