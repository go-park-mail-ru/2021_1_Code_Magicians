package application

import (
	"fmt"
	"io"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type PinApp struct {
	p        repository.PinRepository
	boardApp BoardAppInterface
	s3App    S3AppInterface
}

func NewPinApp(p repository.PinRepository, boardApp BoardAppInterface, s3App S3AppInterface) *PinApp {
	return &PinApp{p, boardApp, s3App}
}

type PinAppInterface interface {
	CreatePin(*entity.Pin) (int, error)
	SavePin(int, int) error
	AddPin(int, int) error             // Saving user's pin
	GetPin(int) (*entity.Pin, error)   // Get pin by pinID
	GetPins(int) ([]entity.Pin, error) // Get pins by boardID
	GetLastUserPinID(int) (int, error)
	SavePicture(*entity.Pin) error
	RemovePin(int, int) error
	DeletePin(int, int) error           // Removes pin by ID
	UploadPicture(int, io.Reader) error // Upload pin
	GetNumOfPins(int) ([]entity.Pin, error)
}

// CreatePin creates passed pin and adds it to native user's board
// It returns pin's assigned ID and nil on success, any number and error on failure
func (pn *PinApp) CreatePin(pin *entity.Pin) (int, error) {
	initBoardID, err := pn.boardApp.GetInitUserBoard(pin.UserID)
	if err != nil {
		return -1, err
	}
	pinID, err := pn.p.CreatePin(pin)
	if err != nil {
		return -1, err
	}

	err = pn.p.AddPin(initBoardID, pinID)
	if err != nil {
		pn.p.DeletePin(pinID)
		return -1, err
	}

	if pin.BoardID != initBoardID && pin.BoardID != 0 {
		err = pn.p.AddPin(pin.BoardID, pinID)
		if err != nil {
			pn.p.DeletePin(pinID)
			return -1, err
		}
	}

	return pinID, nil
}

// SavePin adds any pin to native user's board
// It returns nil on success, error on failure
func (pn *PinApp) SavePin(userID int, pinID int) error {
	initBoardID, err := pn.boardApp.GetInitUserBoard(userID)
	if err != nil {
		return err
	}

	err = pn.p.AddPin(initBoardID, pinID)
	if err != nil {
		return err
	}

	return nil
}

// AddPin adds pin to chosen board
// It returns nil on success, error on failure
func (pn *PinApp) AddPin(boardID int, pinID int) error {
	//board, err := pn.boardApp.GetBoard(boardID)
	//if err != nil {
	//	return err
	//}
	//
	//pin, err := pn.GetPin(pinID)
	//if err != nil {
	//	return err
	//}
	//
	//if board.ImageLink == "" && pin.ImageLink != ""{
	//	err = pn.boardApp.UploadBoardAvatar(boardID, pin.ImageLink)
	//	if err != nil {
	//		return err
	//	}
	//}
	return pn.p.AddPin(boardID, pinID)
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
func (pn *PinApp) DeletePin(boardID int, pinID int) error {
	pin, err := pn.p.GetPin(pinID)
	if err != nil {
		return err
	}

	err = pn.p.RemovePin(boardID, pinID)
	if err != nil {
		return err
	}

	refCount, err := pn.p.PinRefCount(pinID)
	if err != nil {
		return err
	}

	if refCount == 0 {
		err = pn.p.DeletePin(pinID)
		if err != nil {
			return err
		}
		return pn.s3App.DeleteFile(pin.ImageLink)
	}
	return nil
}

// RemovePin deletes pin from user's passed board
// It returns nil on success and error on failure
func (pn *PinApp) RemovePin(boardID int, pinID int) error {
	return pn.p.RemovePin(boardID, pinID)
}

// SavePicture saves path to image of current pin in database
// It returns nil on success and error on failure
func (pn *PinApp) SavePicture(pin *entity.Pin) error {
	return pn.p.SavePicture(pin)
}

// GetLastUserPinID returns path to image of current pin in database
// It returns nil on success and error on failure
func (pn *PinApp) GetLastUserPinID(userID int) (int, error) {
	return pn.p.GetLastUserPinID(userID)
}

//UploadPicture uploads picture to pin and saves new picture path in S3
// It returns nil on success and error on failure
func (pn *PinApp) UploadPicture(pinID int, file io.Reader) error {
	pin, err := pn.GetPin(pinID)
	if err != nil {
		return fmt.Errorf("No pin found to place picture")
	}

	filenamePrefix, err := GenerateRandomString(40) // generating random image
	if err != nil {
		return fmt.Errorf("Could not generate filename")
	}

	picturePath := "pins/" + filenamePrefix + ".jpg"
	err = pn.s3App.UploadFile(file, picturePath)
	if err != nil {
		return fmt.Errorf("File upload failed")
	}

	//initBoardID, err := pn.boardApp.GetInitUserBoard(pin.UserID)
	//if err != nil {
	//	return err
	//}
	//
	//board, err := pn.boardApp.GetBoard(initBoardID)
	//if err != nil {
	//	return err
	//}
	//if board.ImageLink == "" {
	//	err = pn.boardApp.UploadBoardAvatar(initBoardID, picturePath)
	//	if err != nil {
	//		return err
	//	}
	//}

	pin.ImageLink = picturePath

	err = pn.SavePicture(pin)
	if err != nil {
		pn.s3App.DeleteFile(picturePath)
		return fmt.Errorf("Pin saving failed")
	}

	return nil
}

func (pn *PinApp) GetNumOfPins(numOfPins int) ([]entity.Pin, error) {
	return pn.p.GetNumOfPins(numOfPins)
}