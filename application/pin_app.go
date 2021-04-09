package application

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type PinApp struct {
	p repository.PinRepository
}

func NewPinApp(p repository.PinRepository) *PinApp {
	return &PinApp{p}
}

type PinAppInterface interface {
	AddPin(*entity.Pin) (int, error)   // Saving user's pin
	GetPin(int) (*entity.Pin, error)   // Get pin by pinID
	GetPins(int) ([]entity.Pin, error) // Get pins by boardID
	DeletePin(int, int) error               // Removes pin by ID
}

// AddPin adds user's pin to database
// It returns pin's assigned ID and nil on success, any number and error on failure
func (pn *PinApp) AddPin(pin *entity.Pin) (int, error) {
	return pn.p.AddPin(pin)
}

// GetPin returns pin with passed pinID
// It returns that pin and nil on success, nil and error on failure
func (pn *PinApp) GetPin(pinID int) (*entity.Pin, error) {
	return pn.p.GetPin(pinID)
}

// GetPins returns all the pins with passed boardID
// It returns slice of pins and nil on success, nil and error on failure
func (pn *PinApp) GetPins(boardID int) ([]entity.Pin, error) {
	return pn.p.GetPins(boardID)
}

// DeletePin deletes pin with passed pinID
// It returns nil on success and error on failure
func (pn *PinApp) DeletePin(pinID int, userID int) error {
	return pn.p.DeletePin(pinID, userID)
}
