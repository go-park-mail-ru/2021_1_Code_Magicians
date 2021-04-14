package repository

import "pinterest/domain/entity"

type PinRepository interface {
	CreatePin(*entity.Pin) (int, error)
	AddPin(int, int) error   // Saving user's pin
	GetPin(int) (*entity.Pin, error)   // Get pin by pinID
	GetPins(int) ([]entity.Pin, error) // Get pins by boardID
	DeletePin(int) error          // Removes pin by ID
	SavePicture(*entity.Pin) error // Saving picture in database
	PinRefCount(int) (int, error)
	RemovePin(int, int) error
	GetLastUserPinID(int) (int, error)
}
