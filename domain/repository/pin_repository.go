package repository

import "pinterest/domain/entity"

type PinRepository interface {
	SavePin(*entity.Pin) (*entity.Pin, map[string]string)
	GetPin(int) (*entity.Pin, error)   // Get pin by pinID
	GetPins(int) ([]entity.Pin, error) // Get pins by boardID
}
