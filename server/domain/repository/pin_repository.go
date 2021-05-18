package repository

import "pinterest/domain/entity"

type PinRepository interface {
	CreatePin(pin *entity.Pin) (int, error)
	AddPin(boardID int, pinID int) error              // Add pin to specified board
	GetPin(pinID int) (*entity.Pin, error)            // Get pin by pinID
	GetPins(boardID int) ([]entity.Pin, error)        // Get pins by boardID
	DeletePin(pinID int) error                        // Delete pin entirely
	SavePicture(pin *entity.Pin) error                // Update pin's picture properties
	PinRefCount(pinID int) (int, error)               // Get amount of boards pin is in
	RemovePin(boardID int, pinID int) error           // Delete pin from board
	GetLastPinID(userID int) (int, error)             // Get user's last pin's ID
	GetNumOfPins(numOfPins int) ([]entity.Pin, error) // Get specified amount of pins
	SearchPins(keywords string) ([]entity.Pin, error)
	GetPinsByUserID(userID int) ([]entity.Pin, error)   // Get all pins that user added
	GetPinsOfUsers(userIDs []int) ([]entity.Pin, error) // Get all pins belonging to users
}
