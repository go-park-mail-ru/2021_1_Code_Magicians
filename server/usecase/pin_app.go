package usecase

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"pinterest/domain/entity"
	"pinterest/domain/repository"

	"github.com/EdlinOrg/prominentcolor"
)

type PinApp struct {
	p        repository.PinRepository
	boardApp BoardAppInterface
	s3App    S3AppInterface
}

type imageInfo struct {
	height       int
	width        int
	averageColor string
}

func (imageStruct *imageInfo) fillFromImage(imageFile io.Reader) error {
	image, _, err := image.Decode(imageFile)
	if err != nil {
		return fmt.Errorf("Image decoding failed")
	}

	imageStruct.height, imageStruct.width = image.Bounds().Dy(), image.Bounds().Dx()

	colors, err := prominentcolor.Kmeans(image)
	if err != nil {
		return fmt.Errorf("Could not determine image's most prominent color")
	}
	imageStruct.averageColor = colors[0].AsString()

	return nil
}

func NewPinApp(p repository.PinRepository, boardApp BoardAppInterface, s3App S3AppInterface) *PinApp {
	return &PinApp{p, boardApp, s3App}
}

type PinAppInterface interface {
	CreatePin(pin *entity.Pin, file io.Reader, extension string) (int, error)
	SavePin(userID int, pinID int) error                             // Add pin to user's initial board
	AddPin(boardID int, pinID int) error                             // Add pin to specified board
	GetPin(pinID int) (*entity.Pin, error)                           // Get pin by pinID
	GetPins(boardID int) ([]entity.Pin, error)                       // Get pins by boardID
	GetLastPinID(userID int) (int, error)                            // Get user's last pin's ID
	SavePicture(pin *entity.Pin) error                               // Update pin's picture properties
	RemovePin(boardID int, pinID int) error                          // Delete pin from board
	DeletePin(pinID int) error                                       // Delete pin entirely
	UploadPicture(pinID int, file io.Reader, extension string) error // Upload pin's image
	GetNumOfPins(numOfPins int) ([]entity.Pin, error)                // Get specified amount of pins
	SearchPins(keywords string) ([]entity.Pin, error)
}

// CreatePin creates passed pin and adds it to native user's board
// It returns pin's assigned ID and nil on success, any number and error on failure
func (pinApp *PinApp) CreatePin(pin *entity.Pin, file io.Reader, extension string) (int, error) {
	initBoardID, err := pinApp.boardApp.GetInitUserBoard(pin.UserID)
	if err != nil {
		return -1, err
	}
	pinID, err := pinApp.p.CreatePin(pin)
	if err != nil {
		return -1, err
	}

	err = pinApp.p.AddPin(initBoardID, pinID)
	if err != nil {
		pinApp.p.DeletePin(pinID)
		return -1, err
	}

	if pin.BoardID != initBoardID && pin.BoardID != 0 {
		err = pinApp.p.AddPin(pin.BoardID, pinID)
		if err != nil {
			pinApp.p.DeletePin(pinID)
			return -1, err
		}
	}

	err = pinApp.UploadPicture(pinID, file, extension)
	if err != nil {
		return -1, err
	}

	return pinID, nil
}

// SavePin adds any pin to native user's board
// It returns nil on success, error on failure
func (pinApp *PinApp) SavePin(userID int, pinID int) error {
	initBoardID, err := pinApp.boardApp.GetInitUserBoard(userID)
	if err != nil {
		return err
	}

	err = pinApp.p.AddPin(initBoardID, pinID)
	if err != nil {
		return err
	}

	return nil
}

// AddPin adds pin to chosen board
// It returns nil on success, error on failure
func (pinApp *PinApp) AddPin(boardID int, pinID int) error {
	return pinApp.p.AddPin(boardID, pinID)
}

// GetPin returns pin with passed pinID
// It returns that pin and nil on success, nil and error on failure
func (pinApp *PinApp) GetPin(pinID int) (*entity.Pin, error) {
	return pinApp.p.GetPin(pinID)
}

// GetPins returns all the pins with passed boardID
// It returns slice of pins and nil on success, nil and error on failure
func (pinApp *PinApp) GetPins(boardID int) ([]entity.Pin, error) {
	return pinApp.p.GetPins(boardID)
}

// DeletePin deletes pin with passed pinID, deleting associated comments and board relations
// It returns nil on success and error on failure
func (pinApp *PinApp) DeletePin(pinID int) error {
	pin, err := pinApp.p.GetPin(pinID)
	if err != nil {
		return err
	}

	err = pinApp.p.DeletePin(pinID)
	if err != nil {
		return err
	}
	return pinApp.s3App.DeleteFile(pin.ImageLink)
}

// RemovePin deletes pin from user's passed board, deleting pin if no boards reference it
// It returns nil on success and error on failure
func (pinApp *PinApp) RemovePin(boardID int, pinID int) error {
	pin, err := pinApp.p.GetPin(pinID)
	if err != nil {
		return err
	}

	err = pinApp.p.RemovePin(boardID, pinID)
	if err != nil {
		return err
	}

	refCount, err := pinApp.p.PinRefCount(pinID)
	if err != nil {
		return err
	}

	if refCount == 0 {
		err = pinApp.p.DeletePin(pinID)
		if err != nil {
			return err
		}
		return pinApp.s3App.DeleteFile(pin.ImageLink)
	}
	return nil
}

// SavePicture saves path to image of current pin in database
// It returns nil on success and error on failure
func (pinApp *PinApp) SavePicture(pin *entity.Pin) error {
	return pinApp.p.SavePicture(pin)
}

// GetLastUserPinID returns path to image of current pin in database
// It returns nil on success and error on failure
func (pinApp *PinApp) GetLastPinID(userID int) (int, error) {
	return pinApp.p.GetLastPinID(userID)
}

//UploadPicture uploads picture to pin and saves new picture path in S3
// It returns nil on success and error on failure
func (pinApp *PinApp) UploadPicture(pinID int, file io.Reader, extension string) error {
	pin, err := pinApp.GetPin(pinID)
	if err != nil {
		return entity.PinNotFoundError
	}

	fileAsBytes := make([]byte, 0)
	imageStruct := new(imageInfo)
	switch extension {
	case ".png", ".jpg", ".gif":
		fileAsBytes, _ = io.ReadAll(file) // TODO: this may be too slow, rework somehow? Maybe restore file after reading height/width?
		err = imageStruct.fillFromImage(bytes.NewReader(fileAsBytes))
		if err != nil {
			return fmt.Errorf("Image parsing failed")
		}
	default:
		return fmt.Errorf("File extension not supported")
	}

	filenamePrefix, err := GenerateRandomString(40) // generating random filename
	if err != nil {
		return fmt.Errorf("Could not generate filename")
	}

	picturePath := "pins/" + filenamePrefix + extension
	err = pinApp.s3App.UploadFile(bytes.NewReader(fileAsBytes), picturePath)
	if err != nil {
		return fmt.Errorf("File upload failed")
	}

	pin.ImageLink = picturePath
	pin.ImageHeight = imageStruct.height
	pin.ImageWidth = imageStruct.width
	pin.ImageAvgColor = imageStruct.averageColor

	err = pinApp.SavePicture(pin)
	if err != nil {
		pinApp.s3App.DeleteFile(picturePath)
		return fmt.Errorf("Pin saving failed")
	}

	return nil
}

// GetNumOfPins generates the main feed
// It returns numOfPins pins and nil on success, nil and error on failure
func (pinApp *PinApp) GetNumOfPins(numOfPins int) ([]entity.Pin, error) {
	if numOfPins <= 0 {
		return nil, entity.NonPositiveNumOfPinsError
	}
	return pinApp.p.GetNumOfPins(numOfPins)
}

// SearchPins returns pins by keywords
// It returns suitable pins and nil on success, nil and error on failure
func (pinApp *PinApp) SearchPins(keywords string) ([]entity.Pin, error) {
	return pinApp.p.SearchPins(keywords)
}
