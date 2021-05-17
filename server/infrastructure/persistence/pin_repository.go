package persistence

import (
	"context"
	"fmt"
	"pinterest/domain/entity"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PinsRepo struct {
	db *pgxpool.Pool
}

func NewPinsRepository(db *pgxpool.Pool) *PinsRepo {
	return &PinsRepo{db}
}

const createPinQuery string = "INSERT INTO Pins (title, imageLink, imageHeight, imageWidth, ImageAvgColor, description, userID)\n" +
	"values ($1, $2, $3, $4, $5, $6, $7)\n" +
	"RETURNING pinID;\n"

// CreatePin creates new pin with passed fields
// It returns pin's assigned ID and nil on success, any number and error on failure
func (r *PinsRepo) CreatePin(pin *entity.Pin) (int, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), createPinQuery, pin.Title,
		pin.ImageLink, pin.ImageHeight, pin.ImageWidth, pin.ImageAvgColor,
		pin.Description, pin.UserID)
	newPinID := 0
	err = row.Scan(&newPinID)
	if err != nil {
		return -1, entity.CreatePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return newPinID, nil
}

const createPairQuery string = "INSERT INTO pairs (boardID, pinID)\n" +
	"values ($1, $2);\n"

// AddPin add new pin to specified board with passed fields
// It returns nil on success, error on failure
func (r *PinsRepo) AddPin(boardID int, pinID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), createPairQuery, boardID, pinID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.AddPinToBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const deletePinQuery string = "DELETE CASCADE FROM pins WHERE pinID=$1"

// DeletePin deletes pin with passed ID
// It returns nil on success and error on failure
func (r *PinsRepo) DeletePin(pinID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deletePinQuery, pinID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.DeletePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return err
}

const deletePairQuery string = "DELETE FROM pairs WHERE pinID = $1 AND boardID = $2;"

// RemovePin removes pin with passed boardID
// It returns nil on success and error on failure
func (r *PinsRepo) RemovePin(boardID int, pinID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deletePairQuery, pinID, boardID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.RemovePinError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return err
}

const getPinRefCount string = "SELECT COUNT(pinID) FROM pairs WHERE pinID = $1"

// PinRefCount count the number of pin references
// It returns number of references and nil on success and any number and error on failure
func (r *PinsRepo) PinRefCount(pinID int) (int, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	refCount := 0
	row := tx.QueryRow(context.Background(), getPinRefCount, pinID)
	err = row.Scan(&refCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return -1, entity.GetPinReferencesCount
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return refCount, nil
}

const getPinQuery string = "SELECT pinID, userID, title," +
	"imageLink, imageHeight, imageWidth, ImageAvgColor, description\n" +
	"FROM Pins WHERE pinID=$1"

// GetPin fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *PinsRepo) GetPin(pinID int) (*entity.Pin, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	pin := entity.Pin{PinID: pinID}
	row := tx.QueryRow(context.Background(), getPinQuery, pinID)
	err = row.Scan(&pin.PinID, &pin.UserID, &pin.Title,
		&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
		&pin.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.PinNotFoundError
		}
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &pin, nil
}

const getPinsByBoardQuery string = "SELECT pins.pinID, pins.userID, pins.title, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, pins.description\n" +
	"FROM Pins\n" +
	"INNER JOIN pairs on pins.pinID = pairs.pinID WHERE boardID=$1"

// GetPins fetches all pins from board
// It returns slice of all pins in board, nil on success and nil, error on failure
func (r *PinsRepo) GetPins(boardID int) ([]entity.Pin, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	pins := make([]entity.Pin, 0)
	rows, err := tx.Query(context.Background(), getPinsByBoardQuery, boardID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, entity.GetPinsByBoardIdError
	}

	for rows.Next() {
		pin := entity.Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pin.Description)
		if err != nil {
			return nil, err // TODO: error handling
		}
		pins = append(pins, pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return pins, nil
}

const savePictureQuery string = "UPDATE pins\n" +
	"SET imageLink=$1, " +
	"imageHeight=$2, " +
	"imageWidth=$3, " +
	"imageAvgColor=$4\n" +
	"WHERE pinID=$5"

// SavePicture saves pin's picture to database
// It returns nil on success and error on failure
func (r *PinsRepo) SavePicture(pin *entity.Pin) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), savePictureQuery, pin.ImageLink, pin.ImageHeight, pin.ImageWidth, pin.ImageAvgColor, pin.PinID)
	if err != nil {
		// Other errors
		return entity.PinSavingError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const getLastUserPinQuery string = "SELECT pins.pinID\n" +
	"FROM pins\n" +
	"INNER JOIN pairs on pairs.pinID=pins.pinID\n" +
	"INNER JOIN boards on boards.boardID=pairs.boardID AND boards.userID = $1\n" +
	"GROUP BY boards.userID\n" +
	"ORDER BY pins.pinID DESC LIMIT 1\n"

// GetLastPinID
func (r *PinsRepo) GetLastPinID(userID int) (int, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	lastPinID := 0
	row := tx.QueryRow(context.Background(), getLastUserPinQuery, userID)
	err = row.Scan(&lastPinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, fmt.Errorf("Pin not found")
		}
		// Other errors
		return -1, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return lastPinID, nil
}

const getNumOfPinsQuery string = "SELECT pins.pinID, pins.userID, pins.title, " +
	"pins.imageLink, pins.imageHeight, pins.imageWidth, pins.imageAvgColor, pins.description\n" +
	"FROM Pins\n" +
	"LIMIT $1;"

// GetNumOfPins generates the main feed
// It returns numOfPins pins and nil on success, nil and error on failure
func (r *PinsRepo) GetNumOfPins(numOfPins int) ([]entity.Pin, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	pins := make([]entity.Pin, 0)
	rows, err := tx.Query(context.Background(), getNumOfPinsQuery, numOfPins)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	for rows.Next() {
		pin := entity.Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title,
			&pin.ImageLink, &pin.ImageHeight, &pin.ImageWidth, &pin.ImageAvgColor,
			&pin.Description)
		if err != nil {
			return nil, entity.FeedLoadingError
		}
		pins = append(pins, pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return pins, nil
}

const SearchPinsQuery string = "SELECT pins.pinID, pins.userID, pins.title, pins.imageLink, pins.description FROM Pins\n" +
	"WHERE LOWER(pins.title) LIKE $1;"

// SearchPins returns pins by keywords
// It returns suitable pins and nil on success, nil and error on failure
func (r *PinsRepo) SearchPins(keyWords string) ([]entity.Pin, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	pins := make([]entity.Pin, 0)
	rows, err := tx.Query(context.Background(), SearchPinsQuery, "%"+keyWords+"%")
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.NoResultSearch
		}
		return nil, err
	}

	for rows.Next() {
		pin := entity.Pin{}
		err = rows.Scan(&pin.PinID, &pin.UserID, &pin.Title, &pin.ImageLink, &pin.Description)
		if err != nil {
			return nil, entity.SearchingError
		}
		pins = append(pins, pin)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return pins, nil
}
