package pins

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"pinterest/domain/entity"
	. "pinterest/services/pins/proto"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	db *pgxpool.Pool
	s3 *session.Session
}

func NewService(db *pgxpool.Pool, s3 *session.Session) *service {
	return &service{db, s3}
}

const createBoardQuery string = "INSERT INTO Boards (userID, title, description)\n" +
	"values ($1, $2, $3)\n" +
	"RETURNING boardID"
const increaseBoardCountQuery string = "UPDATE Users SET boards_count = boards_count + 1 WHERE userID=$1"

// CreateBoard adds new board with passed fields to database
// It returns board's assigned ID and nil on success, any number and error on failure
func (s *service) CreateBoard(ctx context.Context, board *Board) (*BoardID, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &BoardID{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), createBoardQuery, board.UserID, board.Title, board.Description)
	newBoardID := 0
	err = row.Scan(&newBoardID)
	if err != nil {
		return &BoardID{}, entity.CreateBoardError
	}

	_, err = tx.Exec(context.Background(), increaseBoardCountQuery, board.UserID)
	if err != nil {
		return &BoardID{}, entity.CreateBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &BoardID{}, entity.TransactionCommitError
	}
	return &BoardID{BoardID: int64(newBoardID)}, nil
}

const getBoardQuery string = "SELECT userID, title, description, " +
	"imageLink, imageHeight, imageWidth, imageAvgColor\n" +
	"FROM Boards\n" +
	"WHERE boardID=$1"

// GetBoard fetches board with passed ID from database
// It returns that board, nil on success and nil, error on failure
func (s *service) GetBoard(ctx context.Context, boardID *BoardID) (*Board, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Board{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	board := Board{BoardID: boardID.BoardID}
	row := tx.QueryRow(context.Background(), getBoardQuery, boardID.BoardID)
	err = row.Scan(&board.UserID, &board.Title, &board.Description,
		&board.ImageLink, &board.ImageHeight, &board.ImageWidth, &board.ImageAvgColor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Board{}, entity.BoardNotFoundError
		}
		// Other errors
		return &Board{}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Board{}, entity.TransactionCommitError
	}
	return &board, nil
}

const getBoardsByUserQuery string = "SELECT boardID, title, description, " +
	"imageLink, imageHeight, imageWidth, imageAvgColor\n" +
	"FROM Boards\n" +
	"WHERE userID=$1"

// GetBoards fetches all boards created by user with specified ID from database
// It returns slice of these boards, nil on success and nil, error on failure
func (s *service) GetBoards(ctx context.Context, userID *UserID) (*BoardsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &BoardsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), getBoardsByUserQuery, userID.Uid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &BoardsList{}, entity.GetBoardsByUserIDError
		}
		return &BoardsList{}, err
	}

	boards := make([]*Board, 0)
	for rows.Next() {
		board := Board{UserID: userID.Uid}
		err = rows.Scan(&board.BoardID, &board.Title, &board.Description,
			&board.ImageLink, &board.ImageHeight, &board.ImageWidth, &board.ImageAvgColor)
		if err != nil {
			return &BoardsList{}, err // TODO: error handling
		}
		boards = append(boards, &board)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &BoardsList{}, entity.TransactionCommitError
	}
	return &BoardsList{Boards: boards}, nil
}

const getInitUserBoardQuery string = "SELECT b1.boardID, b1.title, b1.description, " +
	"b1.imageLink, b1.imageHeight, b1.imageWidth, b1.imageAvgColor\n" +
	"FROM boards AS b1\n" +
	"INNER JOIN boards AS b2 on b2.boardID = b1.boardID AND b2.userID = $1\n" +
	"GROUP BY b1.boardID, b2.userID\n" +
	"ORDER BY b2.userID LIMIT 1;"

// GetInitUserBoard gets user's first board from database
// It returns that board and nil on success, nil and error on failure
func (s *service) GetInitUserBoard(ctx context.Context, userID *UserID) (*BoardID, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &BoardID{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	board := Board{UserID: userID.Uid}
	row := tx.QueryRow(context.Background(), getInitUserBoardQuery, userID.Uid)
	err = row.Scan(&board.BoardID, &board.Title, &board.Description,
		&board.ImageLink, &board.ImageHeight, &board.ImageWidth, &board.ImageAvgColor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &BoardID{}, entity.NotFoundInitUserBoard
		}
		return &BoardID{}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &BoardID{}, entity.TransactionCommitError
	}
	return &BoardID{BoardID: board.BoardID}, nil
}

const deleteBoardQuery string = "DELETE FROM Boards WHERE boardID=$1 RETURNING userID"
const decreaseBoardCountQuery string = "UPDATE Users SET boards_count = boards_count - 1 WHERE userID=$1"

// DeleteBoard deletes board with passed id belonging to passed user.
// It returns error if board is not found or if there were problems with database
func (s *service) DeleteBoard(ctx context.Context, boardID *BoardID) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	var boardOwnerID int
	row := tx.QueryRow(context.Background(), deleteBoardQuery, boardID.BoardID)
	err = row.Scan(&boardOwnerID)
	if err != nil {
		return &Error{}, entity.DeleteBoardError
	}

	_, err = tx.Exec(context.Background(), decreaseBoardCountQuery, boardOwnerID)
	if err != nil {
		return &Error{}, entity.DeleteBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, err
}

const saveBoardPictureQuery string = "UPDATE boards\n" +
	"SET imageLink=$2, imageHeight=$3, imageWidth=$4, imageAvgColor=$5\n" +
	"WHERE boardID=$1"

func (s *service) UploadBoardAvatar(ctx context.Context, imageInfo *FileInfo) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), saveBoardPictureQuery, imageInfo.BoardID,
		imageInfo.ImageLink, imageInfo.ImageHeight, imageInfo.ImageWidth, imageInfo.ImageAvgColor)
	if err != nil {
		return &Error{}, err
	}
	if commandTag.RowsAffected() != 1 {
		return &Error{}, entity.BoardAvatarUploadError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const createPinQuery string = "INSERT INTO Pins (userID, title, description, imageLink, imageHeight, imageWidth, imageAvgColor, creationDate)\n" +
	"values ($1, $2, $3, $4, $5, $6, $7, $8)\n" +
	"RETURNING pinID;\n"
const increasePinCountQuery string = "UPDATE Users SET pins_count = pins_count + 1 WHERE userID=$1"

// CreatePin creates new pin with passed fields
// It returns pin's assigned ID and nil on success, any number and error on failure
func (s *service) CreatePin(ctx context.Context, pin *Pin) (*PinID, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &PinID{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), createPinQuery, pin.UserID, pin.Title, pin.Description,
		pin.ImageLink, pin.ImageHeight, pin.ImageWidth, pin.ImageAvgColor,
		pin.CreationDate.AsTime())
	newPinID := 0
	err = row.Scan(&newPinID)
	if err != nil {
		return &PinID{}, entity.CreatePinError
	}

	_, err = tx.Exec(context.Background(), increasePinCountQuery, pin.UserID)
	if err != nil {
		return &PinID{}, entity.CreatePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinID{}, entity.TransactionCommitError
	}
	return &PinID{PinID: int64(newPinID)}, nil
}

const createPairQuery string = "INSERT INTO pairs (boardID, pinID)\n" +
	"values ($1, $2);\n"

// AddPin add new pin to specified board with passed fields
// It returns nil on success, error on failure
func (s *service) AddPin(ctx context.Context, pinInBoard *PinInBoard) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), createPairQuery, pinInBoard.BoardID, pinInBoard.PinID)
	if err != nil {
		return &Error{}, err
	}
	if commandTag.RowsAffected() != 1 {
		return &Error{}, entity.AddPinToBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const getPinQuery string = "SELECT userID, title, description," +
	"imageLink, imageHeight, imageWidth, ImageAvgColor, " +
	"creationDate, reports_count\n" +
	"FROM Pins\n" +
	"WHERE pinID=$1"

// GetPin fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (s *service) GetPin(ctx context.Context, pinID *PinID) (*Pin, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Pin{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), getPinQuery, pinID.PinID)

	pin := Pin{PinID: pinID.PinID}
	var pinCreationDate time.Time
	err = row.Scan(&pin.UserID, &pin.Title, &pin.Description,
		&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
		&pinCreationDate, &pin.ReportsCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Pin{}, entity.PinNotFoundError
		}
		return &Pin{}, err
	}
	pin.CreationDate = timestamppb.New(pinCreationDate)

	err = tx.Commit(context.Background())
	if err != nil {
		return &Pin{}, entity.TransactionCommitError
	}
	return &pin, nil
}

const getPinsByBoardQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM Pins\n" +
	"INNER JOIN pairs on pins.pinID = pairs.pinID WHERE boardID=$1"

// GetPins fetches all pins from board
// It returns slice of all pins in board, nil on success and nil, error on failure
func (s *service) GetPins(ctx context.Context, boardID *BoardID) (*PinsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), getPinsByBoardQuery, boardID.BoardID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &PinsList{}, nil
		}
		return &PinsList{}, err
	}

	pins := make([]*Pin, 0)
	var pinCreationDate time.Time
	for rows.Next() {
		pin := Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.Description,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pinCreationDate, &pin.ReportsCount)
		if err != nil {
			return &PinsList{}, entity.GetPinsByBoardIdError
		}
		pin.CreationDate = timestamppb.New(pinCreationDate)
		pins = append(pins, &pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionCommitError
	}
	return &PinsList{Pins: pins}, nil
}

const getLastUserPinQuery string = "SELECT pins.pinID\n" +
	"FROM pins\n" +
	"INNER JOIN pairs on pairs.pinID=pins.pinID\n" +
	"INNER JOIN boards on boards.boardID=pairs.boardID AND boards.userID = $1\n" +
	"ORDER BY pins.pinID DESC LIMIT 1\n"

// GetLastPinID
func (s *service) GetLastPinID(ctx context.Context, userID *UserID) (*PinID, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &PinID{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	lastPinID := 0
	row := tx.QueryRow(context.Background(), getLastUserPinQuery, userID.Uid)
	err = row.Scan(&lastPinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &PinID{}, entity.PinNotFoundError
		}
		// Other errors
		return &PinID{}, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinID{}, entity.TransactionCommitError
	}
	return &PinID{PinID: int64(lastPinID)}, nil
}

const getLastBoardPinQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM pins\n" +
	"INNER JOIN pairs on pairs.pinID=pins.pinID\n" +
	"INNER JOIN boards on boards.boardID=pairs.boardID AND boards.boardID = $1\n" +
	"ORDER BY pins.pinID DESC LIMIT 1\n"

func (s *service) GetLastBoardPin(ctx context.Context, boardID *BoardID) (*Pin, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Pin{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), getLastBoardPinQuery, boardID.BoardID)

	pin := Pin{}
	var pinCreationDate time.Time
	err = row.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.Description,
		&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
		&pinCreationDate, &pin.ReportsCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Pin{}, entity.PinNotFoundError
		}
		return &Pin{}, err
	}
	pin.CreationDate = timestamppb.New(pinCreationDate)

	err = tx.Commit(context.Background())
	if err != nil {
		return &Pin{}, entity.TransactionCommitError
	}
	return &pin, nil
}

const getBoardsWithPinQuery string = "SELECT boardID, userID, title, description, " +
	"imageLink, imageHeight, imageWidth, imageAvgColor\n" +
	"FROM Boards as board\n" +
	"INNER JOIN pairs on pairs.boardID = board.boardID AND pairs.pinID = $1"

func (s *service) GetBoardsWithPin(ctx context.Context, pinID *PinID) (*BoardsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &BoardsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), getBoardsWithPinQuery, pinID.PinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &BoardsList{}, entity.BoardsNotFoundError
		}
		return &BoardsList{}, err
	}

	boards := make([]*Board, 0)
	for rows.Next() {
		board := Board{}
		err = rows.Scan(&board.BoardID, &board.UserID, &board.Title, &board.Description,
			&board.ImageLink, &board.ImageHeight, &board.ImageWidth, &board.ImageAvgColor)
		if err != nil {
			return &BoardsList{}, err // TODO: error handling
		}
		boards = append(boards, &board)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &BoardsList{}, entity.TransactionCommitError
	}
	return &BoardsList{Boards: boards}, nil
}

const savePictureQuery string = "UPDATE pins\n" +
	"SET imageLink=$1, " +
	"imageHeight=$2, " +
	"imageWidth=$3, " +
	"imageAvgColor=$4\n" +
	"WHERE pinID=$5"

// SavePicture saves pin's picture to database
// It returns nil on success and error on failure
func (s *service) SavePicture(ctx context.Context, pin *Pin) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), savePictureQuery, pin.ImageLink, pin.ImageHeight, pin.ImageWidth, pin.ImageAvgColor, pin.PinID)
	if err != nil {
		// Other errors
		return &Error{}, entity.PinSavingError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const deletePairQuery string = "DELETE FROM pairs WHERE pinID = $1 AND boardID = $2;"

// RemovePin removes pin with passed boardID
// It returns nil on success and error on failure
func (s *service) RemovePin(ctx context.Context, pinInBoard *PinInBoard) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deletePairQuery, pinInBoard.PinID, pinInBoard.BoardID)
	if err != nil {
		return &Error{}, err
	}
	if commandTag.RowsAffected() != 1 {
		return &Error{}, entity.RemovePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const deletePinQuery string = "DELETE CASCADE FROM pins WHERE pinID=$1 RETURNING userID"
const decreasePinCountQuery string = "UPDATE Users SET pins_count = pins_count - 1 WHERE userID=$1"

// DeletePin deletes pin with passed ID
// It returns nil on success and error on failure
func (s *service) DeletePin(ctx context.Context, pinID *PinID) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	var pinOwnerID int
	row := tx.QueryRow(context.Background(), deletePinQuery, pinID.PinID)
	err = row.Scan(&pinOwnerID)
	if err != nil {
		return &Error{}, entity.DeletePinError
	}

	_, err = tx.Exec(context.Background(), decreasePinCountQuery, pinOwnerID)
	if err != nil {
		return &Error{}, entity.DeletePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, err
}

var maxPostAvatarBodySize = 8 * 1024 * 1024 // 8 mB
func (s *service) UploadPicture(stream Pins_UploadPictureServer) error {
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
	newPinPath := "pins/" + filenamePrefix + req.GetExtension() // TODO: pins folder sharding by date

	for {
		req, err = stream.Recv()
		if err == io.EOF {
			log.Print("file recieving is over")
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
		Key:    aws.String(newPinPath),
		Body:   bytes.NewReader(imageData.Bytes()),
	})

	if err != nil {
		return handleS3Error(err)
	}

	res := &UploadImageResponse{
		Path: newPinPath,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return handleS3Error(err)
}

const getNumOfPinsQuery string = "SELECT pins.pinID, pins.userID, pins.title,  pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM Pins\n" +
	"ORDER BY pins.pinID DESC\n" +
	"LIMIT $1;"

// GetNumOfPins generates the main feed
// It returns numOfPins pins and nil on success, nil and error on failure
func (s *service) GetNumOfPins(ctx context.Context, numOfPins *Number) (*PinsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), getNumOfPinsQuery, numOfPins.Number)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &PinsList{}, nil
		}
		return &PinsList{}, err
	}

	pins := make([]*Pin, 0)
	var pinCreationDate time.Time
	for rows.Next() {
		pin := Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.Description,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pinCreationDate, &pin.ReportsCount)
		if err != nil {
			return &PinsList{}, entity.FeedLoadingError
		}
		pin.CreationDate = timestamppb.New(pinCreationDate)
		pins = append(pins, &pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionCommitError
	}
	return &PinsList{Pins: pins}, nil
}

const SearchAllPinsQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM pins\n" +
	"WHERE LOWER(pins.title) LIKE $1;"
const SearchPeriodPinsQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM pins\n" +
	"WHERE LOWER(pins.title) LIKE $1 AND pins.creationdate > now() - interval "

// SearchPins returns pins by keywords
// It returns suitable pins and nil on success, nil and error on failure
func (s *service) SearchPins(ctx context.Context, searchInput *SearchInput) (*PinsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	var rows pgx.Rows

	if searchInput.Date == "all time" {
		rows, err = tx.Query(context.Background(), SearchAllPinsQuery, "%"+searchInput.KeyWords+"%")
	} else if searchInput.Date == "hour" || searchInput.Date == "day" || searchInput.Date == "week" {
		rows, err = tx.Query(context.Background(), SearchPeriodPinsQuery+"'1 "+searchInput.Date+"';", "%"+searchInput.KeyWords+"%")
	} else {
		return &PinsList{}, entity.SearchingError
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return &PinsList{}, nil
		}
		return &PinsList{}, err
	}

	pins := make([]*Pin, 0)
	var pinCreationDate time.Time
	for rows.Next() {
		pin := Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.Description,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pinCreationDate, &pin.ReportsCount)
		if err != nil {
			return &PinsList{}, entity.SearchingError
		}
		pin.CreationDate = timestamppb.New(pinCreationDate)
		pins = append(pins, &pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionCommitError
	}
	return &PinsList{Pins: pins}, nil
}

const GetPinsByUsersIDQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.description, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, " +
	"pins.creationDate, pins.reports_count\n" +
	"FROM Pins\n" +
	"WHERE pins.UserID = ANY($1)" +
	"ORDER BY pins.PinID DESC;" // So that newest pins will come up first

// GetPinsOfUsers outputs all pins of passed users
// It returns slice of pins, nil on success, nil, error on failure
func (s *service) GetPinsOfUsers(ctx context.Context, userIDs *UserIDList) (*PinsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), GetPinsByUsersIDQuery, userIDs)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &PinsList{}, entity.NoResultSearch
		}
		return &PinsList{}, err
	}

	pins := make([]*Pin, 0)
	var pinCreationDate time.Time
	for rows.Next() {
		pin := Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.Description,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pinCreationDate, &pin.ReportsCount)
		if err != nil {
			return &PinsList{}, entity.GetPinsByUserIdError
		}
		pin.CreationDate = timestamppb.New(pinCreationDate)
		pins = append(pins, &pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &PinsList{}, entity.TransactionCommitError
	}
	return &PinsList{Pins: pins}, nil
}

const getPinRefCount string = "SELECT COUNT(pinID) FROM pairs WHERE pinID = $1"

// PinRefCount count the number of pin references
// It returns number of references and nil on success and any number and error on failure
func (s *service) PinRefCount(ctx context.Context, pinID *PinID) (*Number, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Number{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	refCount := 0
	row := tx.QueryRow(context.Background(), getPinRefCount, pinID.PinID)
	err = row.Scan(&refCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Number{}, nil
		}
		return &Number{}, entity.GetPinReferencesCountError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Number{}, entity.TransactionCommitError
	}
	return &Number{Number: int64(refCount)}, nil
}

const createReportQuery string = "INSERT INTO Reports (pinID, senderID, description)\n" +
	"values ($1, $2, $3)\n" +
	"RETURNING reportID"
const increaseReportCountQuery string = "UPDATE Pins SET reports_count = reports_count + 1 WHERE pinID=$1"

// CreateReport adds supplied report to database
// It returns nil on success, error on failure
func (s *service) CreateReport(ctx context.Context, report *Report) (*ReportID, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &ReportID{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), createReportQuery, report.PinID, report.SenderID, report.Description)
	newReportID := 0
	err = row.Scan(&newReportID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Duplicate") {
			return &ReportID{}, entity.DuplicateReportError
		}
		return &ReportID{}, entity.CreateReportError
	}

	_, err = tx.Exec(context.Background(), increaseReportCountQuery, report.PinID)
	if err != nil {
		return &ReportID{}, entity.CreateReportError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &ReportID{}, entity.TransactionCommitError
	}
	return &ReportID{ReportID: int64(newReportID)}, nil
}

func (s *service) DeleteFile(ctx context.Context, filename *FilePath) (*Error, error) {
	deleter := s3.New(s.s3)
	_, err := deleter.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(filename.ImagePath),
	})
	return &Error{}, handleS3Error(err)
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
