package repository

import "pinterest/domain/entity"

type PinRepository interface {
	AddPin(*entity.Pin) (int, error)   // Saving user's pin
	GetPin(int) (*entity.Pin, error)   // Get pin by pinID
	GetPins(int) ([]entity.Pin, error) // Get pins by boardID
	DeletePin(int, int) error          // Removes pin by ID
	SavePicture(pin *entity.Pin) error // Saving picture in database
	GetLastUserPinID(int) (int, error)
}
