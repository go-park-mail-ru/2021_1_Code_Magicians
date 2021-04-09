package persistence

import (
	"context"
	"fmt"
	"pinterest/domain/entity"
	_ "strings"

	"github.com/jackc/pgx/v4"
)

type PinsRepo struct {
	db *pgx.Conn
}

func NewPinsRepository(db *pgx.Conn) *PinsRepo {
	return &PinsRepo{db}
}

const createPinQuery string = "INSERT INTO Pins (boardID, title, imageLink, description)\n" +
	"values ($1, $2, $3, $4)\n" +
	"RETURNING pinID"

// AddPin add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (r *PinsRepo) AddPin(pin *entity.Pin) (int, error) {
	row := r.db.QueryRow(context.Background(), createPinQuery, pin.BoardID, pin.Title, pin.ImageLink, pin.Description)
	newPinID := 0
	err := row.Scan(&newPinID)
	if err != nil {
		// Other errors
		// log.Println(err)
		return -1, err
	}
	return newPinID, nil
}

const deletePinQuery string = "DELETE FROM pins INNER JOIN boards on userID=1$ WHERE pinID=$2"

// SaveUser deletes user with passed ID
// It returns nil on success and error on failure
func (r *PinsRepo) DeletePin(pinID int, userID int) error {
	_, err := r.db.Exec(context.Background(), deletePinQuery, userID, pinID)
	return err
}

const getPinQuery string = "SELECT boardID, title, imageLink, description FROM Pins WHERE pinID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *PinsRepo) GetPin(pinID int) (*entity.Pin, error) {
	pin := entity.Pin{PinId: pinID}
	row := r.db.QueryRow(context.Background(), getPinQuery, pinID)
	err := row.Scan(&pin.BoardID, &pin.Title, &pin.ImageLink, &pin.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No pin found")
		}
		// Other errors
		return nil, err
	}
	return &pin, nil
}

const getPinsByBoardQuery string = "SELECT pinID, boardID, title, imageLink, description FROM Pins WHERE boardID=$1"

// GetUsers fetches all users from database
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
		err := rows.Scan(&pin.PinId, &pin.BoardID, &pin.Title, &pin.ImageLink, &pin.Description)
		if err != nil {
			return nil, err // TODO: error handling
		}
		pins = append(pins, pin)
	}
	return pins, nil
}