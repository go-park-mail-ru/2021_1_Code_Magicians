package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"pinterest/domain/entity"
	. "pinterest/services/user/proto"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	db *pgxpool.Pool
	s3 *session.Session
}

func NewService(db *pgxpool.Pool, s3 *session.Session) *service {
	return &service{db, s3}
}

const createUserQueryDefaulAvatar string = "INSERT INTO Users (username, passwordhash, salt, email, first_name, last_name, vk_id, avatar)\n" +
	"values ($1, $2, $3, $4, $5, $6, $7, DEFAULT)\n" +
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

	row := tx.QueryRow(context.Background(), createUserQueryDefaulAvatar,
		user.Username, user.Password, user.Salt, user.Email, &firstNamePtr, &lastNamePtr, user.VkID)

	newUserID := 0
	err = row.Scan(&newUserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return nil, entity.UsernameEmailDuplicateError
		}
		return nil, entity.UserScanError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}

	return &UserID{Uid: int64(newUserID)}, nil
}

const saveUserQuery string = "UPDATE Users\n" +
	"SET username=$1, email=$2, first_name=$3, last_name=$4, avatar=$5, vk_id=$6\n" +
	"WHERE userID=$7"

func (s *service) SaveUser(ctx context.Context, us *UserEditInput) (*Error, error) {
	user := FillFromEditForm(us)
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), saveUserQuery, user.Username, user.Email,
		user.FirstName, user.LastName, user.Avatar, user.VkID, user.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return &Error{}, entity.UsernameEmailDuplicateError
		}
		return &Error{}, entity.UserScanError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

var maxPostAvatarBodySize = 8 * 1024 * 1024 // 8 mB
func (s *service) UpdateAvatar(stream User_UpdateAvatarServer) error {
	imageData := bytes.Buffer{}
	imageSize := 0
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot receive image info")
	}

	filenamePrefix, err := entity.GenerateRandomString(40) // generating random filename
	if err != nil {
		return entity.FilenameGenerationError
	}
	newAvatarPath := "avatars/" + filenamePrefix + req.GetExtension() // TODO: avatars folder sharding by date

	for {
		req, err = stream.Recv()
		if err == io.EOF {
			log.Print("file receiving is over")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}
		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > maxPostAvatarBodySize {
			return status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxPostAvatarBodySize)
		}
		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}
	}
	uploader := s3manager.NewUploader(s.s3)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		ACL:    aws.String("public-read"),
		Key:    aws.String(newAvatarPath),
		Body:   bytes.NewReader(imageData.Bytes()),
	})
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot upload to s3: %v", err)
	}

	res := &UploadAvatarResponse{
		Path: newAvatarPath,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return handleS3Error(err)
}

func (s *service) DeleteFile(ctx context.Context, filename *FilePath) (*Error, error) {
	deleter := s3.New(s.s3)
	_, err := deleter.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(filename.ImagePath),
	})
	return &Error{}, handleS3Error(err)
}

const changePasswordQuery string = "UPDATE Users\n" +
	"SET passwordhash=$1\n" +
	"WHERE userID=$2"

func (s *service) ChangePassword(ctx context.Context, pswrd *Password) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), changePasswordQuery, pswrd.Password, pswrd.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return &Error{}, entity.UsernameEmailDuplicateError
		}
		return &Error{}, entity.UserScanError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const deleteUserQuery string = "DELETE FROM Users WHERE userID=$1"

// DeleteUser deletes user with passed ID
// It returns nil on success and error on failure
func (s *service) DeleteUser(ctx context.Context, userID *UserID) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deleteUserQuery, userID.Uid)
	if err != nil {
		return &Error{}, err
	}
	if commandTag.RowsAffected() != 1 {
		return &Error{}, entity.UserNotFoundError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const getUserQuery string = "SELECT username, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
	"FROM Users WHERE userID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (s *service) GetUser(ctx context.Context, userID *UserID) (*UserOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	user := UserOutput{UserID: userID.Uid}
	firstNamePtr := new(string)
	secondNamePtr := new(string)
	avatarPtr := new(string)

	row := tx.QueryRow(context.Background(), getUserQuery, userID.Uid)
	err = row.Scan(&user.Username, &user.Email, &firstNamePtr,
		&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
		&user.BoardsCount, &user.PinsCount, &user.VkID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}
		return nil, entity.UserScanError
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

const getUsersQuery string = "SELECT userID, username, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
	"FROM Users"

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (s *service) GetUsers(ctx context.Context, in *empty.Empty) (*UsersListOutput, error) {
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
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
			&user.BoardsCount, &user.PinsCount, &user.VkID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, entity.UserNotFoundError
			}
			return nil, entity.UserScanError
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

const getUserByUsernameQuery string = "SELECT userID, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
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

	row := tx.QueryRow(context.Background(), getUserByUsernameQuery, username.Username)
	err = row.Scan(&user.UserID, &user.Email, &firstNamePtr,
		&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
		&user.BoardsCount, &user.PinsCount, &user.VkID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserNotFoundError
		}
		return nil, entity.UserScanError
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
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	_, err = tx.Exec(context.Background(), followQuery, follows.FollowerID, follows.FollowedID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return &Error{}, entity.FollowAlreadyExistsError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_followed"`) {
			return &Error{}, entity.UserNotFoundError
		}
		if strings.Contains(err.Error(), `violates foreign key constraint "followers_users_follower"`) { // Actually does not usually happen because of checks in middleware
			return &Error{}, entity.UserNotFoundError
		}

		return &Error{}, err
	}

	_, err = tx.Exec(context.Background(), updateFollowingQuery, follows.FollowerID)
	if err != nil {
		return &Error{}, entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateFollowedByQuery, follows.FollowedID)
	if err != nil {
		return &Error{}, entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const unfollowQuery string = "DELETE FROM Followers WHERE followerID=$1 AND followedID=$2"
const updateUnfollowingQuery string = "UPDATE Users SET following = following - 1 WHERE userID=$1"
const updateUnfollowedByQuery string = "UPDATE Users SET followed_by = followed_by - 1 WHERE userID=$1"

func (s *service) Unfollow(ctx context.Context, follows *Follows) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background()) // Will help if one of updateX queries fails

	result, _ := tx.Exec(context.Background(), unfollowQuery, follows.FollowerID, follows.FollowedID)

	if result.RowsAffected() != 1 {
		return &Error{}, entity.FollowNotFoundError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowingQuery, follows.FollowerID)
	if err != nil {
		return &Error{}, entity.FollowCountUpdateError
	}

	_, err = tx.Exec(context.Background(), updateUnfollowedByQuery, follows.FollowedID)
	if err != nil {
		return &Error{}, entity.FollowCountUpdateError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, err
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

		return &IfFollowedResponse{IsFollowed: false}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &IfFollowedResponse{IsFollowed: false}, entity.TransactionCommitError
	}
	return &IfFollowedResponse{IsFollowed: true}, nil
}

const SearchUsersQuery string = "SELECT userID, username, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
	"FROM Users\n" +
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
	rows, err := tx.Query(context.Background(), SearchUsersQuery, "%"+keyWords.KeyWords+"%")
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UsersNotFoundError
		}
		return nil, err
	}

	for rows.Next() {
		user := UserOutput{}

		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
			&user.BoardsCount, &user.PinsCount, &user.VkID)
		if err != nil {
			return nil, entity.UserScanError
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

const getAllFollowersQuery = "SELECT userID, username, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
	"FROM Users\n" +
	"INNER JOIN (SELECT * FROM Followers WHERE followedID = $1) as users_followers\n" +
	"ON followerID = userID"

// GetAllFollowers fetches all users that follow user with passed ID
// It returns slice of users, nil on success, nil, error on failure
// ! No followers found also counts as an error, entity.UsersNotFoundError
func (s *service) GetAllFollowers(ctx context.Context, userID *UserID) (*UsersListOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	followers := make([]*UserOutput, 0)
	rows, err := tx.Query(context.Background(), getAllFollowersQuery, userID.Uid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UsersNotFoundError
		}
		return nil, err
	}

	for rows.Next() {
		user := UserOutput{}
		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
			&user.BoardsCount, &user.PinsCount, &user.VkID)
		if err != nil {
			return nil, entity.UserScanError
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		followers = append(followers, &user)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}

	return &UsersListOutput{Users: followers}, nil
}

const getAllFollowedQuery = "SELECT userID, username, email, first_name, last_name, avatar, " +
	"followed_by, following, boards_count, pins_count, vk_id\n" +
	"FROM Users\n" +
	"INNER JOIN (SELECT * FROM Followers WHERE followerID = $1) as users_followed\n" +
	"ON followedID = userID"

// GetAllFollowed fetches all users that are followed by user with passed ID
// It returns slice of users, nil on success, nil, error on failure
// ! No followers found also counts as an error, entity.UsersNotFoundError
func (s *service) GetAllFollowed(ctx context.Context, userID *UserID) (*UsersListOutput, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	followed := make([]*UserOutput, 0)
	rows, err := tx.Query(context.Background(), getAllFollowedQuery, userID.Uid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UsersNotFoundError
		}
		return nil, err
	}

	for rows.Next() {
		user := UserOutput{}
		firstNamePtr := new(string)
		secondNamePtr := new(string)
		avatarPtr := new(string)

		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &firstNamePtr,
			&secondNamePtr, &avatarPtr, &user.FollowedBy, &user.Following,
			&user.BoardsCount, &user.PinsCount, &user.VkID)
		if err != nil {
			return nil, entity.UserScanError
		}

		user.FirstName = *emptyIfNil(firstNamePtr)
		user.LastName = *emptyIfNil(secondNamePtr)
		user.Avatar = *emptyIfNil(avatarPtr)
		followed = append(followed, &user)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}

	return &UsersListOutput{Users: followed}, nil
}

func FillFromRegForm(us *UserReg) entity.User {
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

func FillFromEditForm(us *UserEditInput) *entity.User {
	return &entity.User{
		UserID:     int(us.UserID),
		Username:   us.Username,
		Password:   us.Password,
		FirstName:  us.FirstName,
		LastName:   us.LastName,
		Email:      us.Email,
		Avatar:     us.AvatarLink,
		Salt:       us.Salt, // TODO salt realize
		Following:  0,
		FollowedBy: 0,
		VkID:       int(us.VkID),
	}
}

// emptyIfNil replaces nil input with pointer to empty string, noop otherwise
func emptyIfNil(input *string) *string {
	if input == nil {
		return new(string)
	}
	return input
}

func handleS3Error(err error) error {
	if err == nil {
		return nil
	}

	aerr, ok := err.(awserr.Error)
	if ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchBucket:
			return fmt.Errorf("Specified bucket does not exist")
		case s3.ErrCodeNoSuchKey:
			return fmt.Errorf("No file found with such filename")
		case s3.ErrCodeObjectAlreadyInActiveTierError:
			return fmt.Errorf("S3 bucket denied access to you")
		default:
			return fmt.Errorf("Unknown S3 error")
		}
	}

	return fmt.Errorf("Not an S3 error")
}
