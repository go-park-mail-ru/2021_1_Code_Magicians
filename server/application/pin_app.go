package application

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"pinterest/domain/entity"
	grpcPins "pinterest/services/pins/proto"
	"strings"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PinApp struct {
	grpcClient grpcPins.PinsClient
	boardApp   BoardAppInterface
}

type imageInfo struct {
	height       int
	width        int
	averageColor string
}

func NewPinApp(grpcClient grpcPins.PinsClient, boardApp BoardAppInterface) *PinApp {
	return &PinApp{grpcClient, boardApp}
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
	GetPinsWithOffset(offset int, amount int) ([]entity.Pin, error)  // Get specified amount of pins
	SearchPins(keywords string, date string) ([]entity.Pin, error)
	GetPinsOfUsers(userIDs []int) ([]entity.Pin, error) // Get all pins belonging to users
	CreateReport(report *entity.Report) (int, error)
}

// CreatePin creates passed pin and adds it to native user's board
// It returns pin's assigned ID and nil on success, any number and error on failure
func (pinApp *PinApp) CreatePin(pin *entity.Pin, file io.Reader, extension string) (int, error) {
	if pin.BoardID == 0 { // If board was not specified, add pin to default board
		var err error
		pin.BoardID, err = pinApp.boardApp.GetInitUserBoard(pin.UserID)
		if err != nil {
			return -1, err
		}
	}

	pin.CreationDate = time.Now()

	grpcPin := grpcPins.Pin{}
	ConvertToGrpcPin(&grpcPin, pin)
	pinID, err := pinApp.grpcClient.CreatePin(context.Background(), &grpcPin)
	if err != nil {
		return -1, err
	}

	err = pinApp.UploadPicture(int(pinID.PinID), file, extension)
	if err != nil {
		pinApp.grpcClient.DeletePin(context.Background(), pinID)
		return -1, err
	}

	err = pinApp.AddPin(pin.BoardID, int(pinID.PinID))
	if err != nil {
		pinApp.grpcClient.DeletePin(context.Background(), pinID)
		pinApp.grpcClient.DeleteFile(context.Background(), &grpcPins.FilePath{ImagePath: pin.ImageLink})
		if strings.Contains(err.Error(), entity.AddPinToBoardError.Error()) {
			return -1, entity.AddPinToBoardError
		}
		return -1, err
	}

	return int(pinID.PinID), nil
}

// SavePin adds any pin to native user's board
// It returns nil on success, error on failure
func (pinApp *PinApp) SavePin(userID int, pinID int) error {
	initBoardID, err := pinApp.boardApp.GetInitUserBoard(userID)
	if err != nil {
		return err
	}

	err = pinApp.AddPin(initBoardID, pinID)
	if err != nil {
		return err
	}

	return nil
}

// AddPin adds pin to chosen board
// It returns nil on success, error on failure
func (pinApp *PinApp) AddPin(boardID int, pinID int) error {
	pin, err := pinApp.GetPin(pinID)
	if err != nil {
		return err
	}

	_, err = pinApp.grpcClient.AddPin(context.Background(), &grpcPins.PinInBoard{
		BoardID: int64(boardID), PinID: int64(pinID),
	})
	if err != nil {
		if strings.Contains(err.Error(), entity.AddPinToBoardError.Error()) {
			return entity.AddPinToBoardError
		}
		return err
	}

	avatarInfo := new(grpcPins.FileInfo)
	avatarInfo.BoardID = int64(boardID)
	avatarInfo.ImageLink = pin.ImageLink
	avatarInfo.ImageHeight = int64(pin.ImageHeight)
	avatarInfo.ImageWidth = int64(pin.ImageWidth)
	avatarInfo.ImageAvgColor = pin.ImageAvgColor
	_, err = pinApp.grpcClient.UploadBoardAvatar(context.Background(), avatarInfo)
	if err != nil {
		if strings.Contains(err.Error(), entity.BoardAvatarUploadError.Error()) {
			return entity.BoardAvatarUploadError
		}
		return err
	}
	return nil
}

// GetPin returns pin with passed pinID
// It returns that pin and nil on success, nil and error on failure
func (pinApp *PinApp) GetPin(pinID int) (*entity.Pin, error) {
	grpcPin, err := pinApp.grpcClient.GetPin(context.Background(), &grpcPins.PinID{PinID: int64(pinID)})
	if err != nil {
		return nil, err
	}
	pin := entity.Pin{}
	ConvertFromGrpcPin(&pin, grpcPin)
	return &pin, nil
}

// GetPins returns all the pins with passed boardID
// It returns slice of pins and nil on success, nil and error on failure
func (pinApp *PinApp) GetPins(boardID int) ([]entity.Pin, error) {
	grpcPinsList, err := pinApp.grpcClient.GetPins(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	if err != nil {
		return nil, err
	}
	return ConvertGrpcPins(grpcPinsList), nil
}

// DeletePin deletes pin with passed pinID, deleting associated comments and board relations
// It returns nil on success and error on failure
func (pinApp *PinApp) DeletePin(pinID int) error {
	pin, err := pinApp.GetPin(pinID)
	if err != nil {
		return err
	}

	boards, err := pinApp.grpcClient.GetBoardsWithPin(context.Background(), &grpcPins.PinID{PinID: int64(pinID)})
	switch {
	case err == nil:
		for _, board := range boards.Boards {
			pinApp.RemovePin(int(board.BoardID), pin.PinID)
		}
	case strings.Contains(err.Error(), entity.BoardsNotFoundError.Error()):
		break
	default:
		return err
	}

	_, err = pinApp.grpcClient.DeletePin(context.Background(), &grpcPins.PinID{PinID: int64(pinID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.DeletePinError.Error()) {
			return entity.DeletePinError
		}
		return err
	}

	_, err = pinApp.grpcClient.DeleteFile(context.Background(), &grpcPins.FilePath{ImagePath: pin.ImageLink})
	if err != nil {
		return entity.FileDeletionError
	}

	return nil
}

// RemovePin deletes pin from user's passed board
// It returns nil on success and error on failure
func (pinApp *PinApp) RemovePin(boardID int, pinID int) error {
	pin, err := pinApp.GetPin(pinID)
	if err != nil {
		return err
	}

	_, err = pinApp.grpcClient.RemovePin(context.Background(), &grpcPins.PinInBoard{
		BoardID: int64(boardID), PinID: int64(pinID),
	})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.RemovePinError.Error()):
			return entity.RemovePinError
		case strings.Contains(err.Error(), entity.DeletePinError.Error()):
			return entity.DeletePinError
		default:
			return err
		}
	}

	lastPin, err := pinApp.grpcClient.GetLastBoardPin(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	var boardAvatar = new(grpcPins.FileInfo)
	boardAvatar.BoardID = int64(boardID)
	switch {
	case err == nil:
		boardAvatar.ImageLink = lastPin.ImageLink
		boardAvatar.ImageHeight = int64(lastPin.ImageHeight)
		boardAvatar.ImageWidth = int64(lastPin.ImageWidth)
		boardAvatar.ImageAvgColor = lastPin.ImageAvgColor
	case strings.Contains(err.Error(), entity.PinNotFoundError.Error()): // If there are no pins left, we take default image
		boardAvatar.ImageLink = string(entity.BoardAvatarDefaultPath)
		boardAvatar.ImageHeight = 480
		boardAvatar.ImageWidth = 1200
		boardAvatar.ImageAvgColor = "5a5a5a"
	default:
		return err
	}

	_, err = pinApp.grpcClient.UploadBoardAvatar(context.Background(), boardAvatar)
	if err != nil {
		return err
	}

	refCount, err := pinApp.grpcClient.PinRefCount(context.Background(), &grpcPins.PinID{PinID: int64(pinID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.GetPinReferencesCountError.Error()) {
			return entity.GetPinReferencesCountError
		}
		return err
	}

	if refCount.Number == 0 {
		_, err = pinApp.grpcClient.DeletePin(context.Background(), &grpcPins.PinID{PinID: int64(pinID)})
		if err != nil {
			if strings.Contains(err.Error(), entity.DeletePinError.Error()) {
				return entity.DeletePinError
			}
			return err
		}
		_, err = pinApp.grpcClient.DeleteFile(context.Background(), &grpcPins.FilePath{ImagePath: pin.ImageLink})
		return err // S3 errors are not handled in any special way, they all cause InternalServerError
	}

	return nil
}

// SavePicture saves path to image of current pin in database
// It returns nil on success and error on failure
func (pinApp *PinApp) SavePicture(pin *entity.Pin) error {
	grpcPin := grpcPins.Pin{}
	ConvertToGrpcPin(&grpcPin, pin)
	_, err := pinApp.grpcClient.SavePicture(context.Background(), &grpcPin)
	if err != nil {
		if strings.Contains(err.Error(), entity.PinSavingError.Error()) {
			return entity.PinSavingError
		}
		return err
	}

	return nil
}

// GetLastPinID returns path to image of current pin in database
// It returns nil on success and error on failure
func (pinApp *PinApp) GetLastPinID(userID int) (int, error) {
	grpcPinID, err := pinApp.grpcClient.GetLastPinID(context.Background(), &grpcPins.UserID{Uid: int64(userID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.PinNotFoundError.Error()) {
			return -1, entity.PinNotFoundError
		}
		return -1, err
	}

	return int(grpcPinID.PinID), err
}

//UploadPicture uploads picture to pin and saves new picture path in S3
// It returns nil on success and error on failure
func (pinApp *PinApp) UploadPicture(pinID int, file io.Reader, extension string) error {
	pin, err := pinApp.GetPin(pinID)
	if err != nil {
		return entity.PinNotFoundError
	}

	var fileAsBytes []byte
	imageStruct := new(imageInfo)
	switch extension {
	case ".png", ".jpg", ".gif", ".jpeg":
		fileAsBytes, _ = io.ReadAll(file) // TODO: this may be too slow, rework somehow? Maybe restore file after reading height/width?
		err = imageStruct.fillFromImage(bytes.NewReader(fileAsBytes))
		if err != nil {
			return fmt.Errorf("Image parsing failed")
		}
	default:
		return fmt.Errorf("File extension not supported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := pinApp.grpcClient.UploadPicture(ctx)
	for err != nil {
		switch {
		case strings.Contains(err.Error(), entity.FilenameGenerationError.Error()):
			stream, err = pinApp.grpcClient.UploadPicture(ctx)
		default:
			return entity.FileUploadError
		}
	}

	req := &grpcPins.UploadImage{
		Data: &grpcPins.UploadImage_Extension{
			Extension: extension,
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}
	reader := bytes.NewReader(fileAsBytes)
	buffer := make([]byte, 8*1024*1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req = &grpcPins.UploadImage{
			Data: &grpcPins.UploadImage_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	pin.ImageLink = res.Path
	pin.ImageHeight = imageStruct.height
	pin.ImageWidth = imageStruct.width
	pin.ImageAvgColor = imageStruct.averageColor

	err = pinApp.SavePicture(pin)
	if err != nil {
		pinApp.grpcClient.DeleteFile(context.Background(), &grpcPins.FilePath{ImagePath: res.Path})

		if strings.Contains(err.Error(), entity.PinSavingError.Error()) {
			return entity.PinSavingError
		}
		return err
	}

	return nil
}

// GetPinsWithOffset generates the main feed
// It returns ~amount pins and nil on success, nil and error on failure
func (pinApp *PinApp) GetPinsWithOffset(offset int, amount int) ([]entity.Pin, error) {
	grpcPinsList, err := pinApp.grpcClient.GetPinsWithOffset(
		context.Background(),
		&grpcPins.FeedInfo{Offset: int64(offset), Amount: int64(amount)},
	)
	if err != nil {
		if strings.Contains(err.Error(), entity.FeedLoadingError.Error()) {
			return nil, entity.FeedLoadingError
		}
		return nil, err
	}

	return ConvertGrpcPins(grpcPinsList), nil
}

// SearchPins returns pins by keywords
// It returns suitable pins and nil on success, nil and error on failure
func (pinApp *PinApp) SearchPins(keyWords string, date string) ([]entity.Pin, error) {
	grpcPinsList, err := pinApp.grpcClient.SearchPins(context.Background(),
		&grpcPins.SearchInput{KeyWords: keyWords, Date: date})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.NoResultSearch.Error()):
			return nil, entity.NoResultSearch
		case strings.Contains(err.Error(), entity.SearchingError.Error()):
			return nil, entity.SearchingError
		default:
			return nil, err
		}
	}

	return ConvertGrpcPins(grpcPinsList), nil
}

// GetPinsOfUsers outputs all pins of passed users
// It returns slice of pins, nil on success, nil, error on failure
func (pinApp *PinApp) GetPinsOfUsers(userIDs []int) ([]entity.Pin, error) {
	userIdsForGrpc := make([]int64, 0, len(userIDs))
	for _, userID := range userIDs {
		userIdsForGrpc = append(userIdsForGrpc, int64(userID))
	}

	grpcPinsList, err := pinApp.grpcClient.GetPinsOfUsers(context.Background(), &grpcPins.UserIDList{Ids: userIdsForGrpc})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.NoResultSearch.Error()):
			return nil, entity.NoResultSearch
		case strings.Contains(err.Error(), entity.GetPinsByUserIdError.Error()):
			return nil, entity.GetPinsByUserIdError
		default:
			return nil, err
		}
	}

	return ConvertGrpcPins(grpcPinsList), nil
}

// CreateReport adds report with parameters of passed report struct to database
// It returns added report's ID, nil on success, -1, error on failure
func (pinApp *PinApp) CreateReport(report *entity.Report) (int, error) {
	_, err := pinApp.GetPin(report.PinID)
	if err != nil {
		return -1, err
	}

	grpcReport := grpcPins.Report{}
	grpcReport.PinID = int64(report.PinID)
	grpcReport.SenderID = int64(report.SenderID)
	grpcReport.Description = report.Description

	grpcReportID, err := pinApp.grpcClient.CreateReport(context.Background(), &grpcReport)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.CreateReportError.Error()):
			return -1, entity.CreateReportError
		case strings.Contains(err.Error(), entity.DuplicateReportError.Error()):
			return -1, entity.DuplicateReportError
		default:
			return -1, err
		}
	}
	return int(grpcReportID.ReportID), nil
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

func ConvertToGrpcPin(grpcPin *grpcPins.Pin, pin *entity.Pin) {
	grpcPin.UserID = int64(pin.UserID)
	grpcPin.PinID = int64(pin.PinID)
	grpcPin.BoardID = int64(pin.BoardID)
	grpcPin.Title = pin.Title
	grpcPin.Description = pin.Description
	grpcPin.ImageAvgColor = pin.ImageAvgColor
	grpcPin.ImageWidth = int64(pin.ImageWidth)
	grpcPin.ImageHeight = int64(pin.ImageHeight)
	grpcPin.ImageLink = pin.ImageLink
	grpcPin.CreationDate = timestamppb.New(pin.CreationDate)
	grpcPin.ReportsCount = int64(pin.ReportsCount)
}

func ConvertFromGrpcPin(pin *entity.Pin, grpcPin *grpcPins.Pin) {
	pin.UserID = int(grpcPin.UserID)
	pin.PinID = int(grpcPin.PinID)
	pin.BoardID = int(grpcPin.BoardID)
	pin.Title = grpcPin.Title
	pin.Description = grpcPin.Description
	pin.ImageAvgColor = grpcPin.ImageAvgColor
	pin.ImageWidth = int(grpcPin.ImageWidth)
	pin.ImageHeight = int(grpcPin.ImageHeight)
	pin.ImageLink = grpcPin.ImageLink
	pin.CreationDate = grpcPin.CreationDate.AsTime()
	pin.ReportsCount = int(grpcPin.ReportsCount)
}

func ConvertGrpcPins(grpcPins *grpcPins.PinsList) []entity.Pin {
	pins := make([]entity.Pin, 0)
	for _, grpcPin := range grpcPins.Pins {
		pin := entity.Pin{}
		ConvertFromGrpcPin(&pin, grpcPin)
		pins = append(pins, pin)
	}
	return pins
}
