package persistence

import (
	"context"
	"errors"
	"fmt"
	"pinterest/domain/entity"

	"github.com/jackc/pgx/v4"
)

type PinsRepo struct {
	db *pgx.Conn
}

func NewPinsRepository(db *pgx.Conn) *PinsRepo {
	return &PinsRepo{db}
}

const createPinQuery string = "INSERT INTO Pins (title, imageLink, description)\n" +
	"values ($1, $2, $3)\n" +
	"RETURNING pinID;\n"

// CreatePin creates new pin with passed fields
// It returns pin's assigned ID and nil on success, any number and error on failure
func (r *PinsRepo) CreatePin(pin *entity.Pin) (int, error) {
	row := r.db.QueryRow(context.Background(), createPinQuery, pin.Title, pin.ImageLink, pin.Description)
	newPinID := 0
	err := row.Scan(&newPinID)
	if err != nil {
		return -1, err
	}

	return newPinID, nil
}

const createPairQuery string = "INSERT INTO pairs (boardID, pinID)\n" +
	"values ($1, $2);\n"

// AddPin add new pin to specified board with passed fields
// It returns nil on success, error on failure
func (r *PinsRepo) AddPin(boardID int, pinID int) error {
	commandTag, err := r.db.Exec(context.Background(), createPairQuery, boardID, pinID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("Pin not found")
	}

	return nil
}

const deletePinQuery string = "DELETE FROM pins WHERE pinID=$2"

// DeletePin deletes pin with passed ID
// It returns nil on success and error on failure
func (r *PinsRepo) DeletePin(pinID int, userID int) error {
	commandTag, err := r.db.Exec(context.Background(), deletePinQuery, userID, pinID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("Pin not found")
	}
	return err
}

const getPinQuery string = "SELECT boardID, title, imageLink, description FROM Pins WHERE pinID=$1"

// GetPin fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *PinsRepo) GetPin(pinID int) (*entity.Pin, error) {
	pin := entity.Pin{PinId: pinID}
	row := r.db.QueryRow(context.Background(), getPinQuery, pinID)
	err := row.Scan(&pin.Title, &pin.ImageLink, &pin.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("Pin not found")
		}
		// Other errors
		return nil, err
	}
	return &pin, nil
}

const getPinsByBoardQuery string = "SELECT pins.pinID, pins.title, pins.imageLink, pins.description FROM Pins\n" +
	"INNER JOIN pairs on pins.pinID = pairs.pinID WHERE boardID=$1"

// GetPins fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (r *PinsRepo) GetPins(boardID int) ([]entity.Pin, error) {
	pins := make([]entity.Pin, 0)
	rows, err := r.db.Query(context.Background(), getPinsByBoardQuery, boardID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No pins found")
		}

		// Other errors
		return nil, err
	}

	for rows.Next() {
		pin := entity.Pin{}
		err := rows.Scan(&pin.PinId, &pin.Title, &pin.ImageLink, &pin.Description)
		if err != nil {
			return nil, err // TODO: error handling
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

const savePictureQuery string = "UPDATE pins\n" +
	"SET imageLink=$1\n" +
	"WHERE pinID=$2"

// SavePicture saves pin's picture to database
// It returns nil on success and error on failure
func (r *PinsRepo) SavePicture(pin *entity.Pin) error {
	_, err := r.db.Exec(context.Background(), savePictureQuery, pin.ImageLink, pin.PinId)
	if err != nil {
		// Other errors
		return err
	}
	return nil
}

const getLastUserPinQuery string = "SELECT pins.pinID\n" +
	"FROM pins\n" +
	"INNER JOIN pairs on pairs.pinID=pins.pinID\n" +
	"INNER JOIN boards on boards.boardID=pairs.boardID AND boards.userID = $1\n" +
	"GROUP BY boards.userID\n" +
	"ORDER BY pins.pinID DESC LIMIT 1\n"

// GetLastUserPinId - ???????????????????????????????????????????????????????????????????????????????????????????????????????????????????
func (r *PinsRepo) GetLastUserPinID(userID int) (int, error) {
	lastPinID := 0
	row := r.db.QueryRow(context.Background(), getLastUserPinQuery, userID)
	err := row.Scan(&lastPinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, fmt.Errorf("Pin not found")
		}
		// Other errors
		return -1, err
	}
	return lastPinID, nil
}
