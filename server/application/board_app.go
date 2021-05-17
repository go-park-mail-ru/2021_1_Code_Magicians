package application

import (
	"context"
	"pinterest/domain/entity"
	grpcPins "pinterest/services/pins/proto"
	"strings"
)

type BoardApp struct {
	grpcClient grpcPins.PinsClient
}

func NewBoardApp(grpcClient grpcPins.PinsClient) *BoardApp {
	return &BoardApp{grpcClient}
}

type BoardAppInterface interface {
	AddBoard(board *entity.Board) (int, error)       // Creating user's board
	GetBoard(boardID int) (*entity.BoardInfo, error) // Get description of the board
	GetBoards(userID int) ([]entity.Board, error)    // Get boards by authorID
	GetInitUserBoard(userID int) (int, error)
	DeleteBoard(userID int, boardID int) error // Removes user's board by ID
	CheckBoard(userID int, boardID int) error
	UploadBoardAvatar(boardID int, avatarLink string) error
}

// AddBoard adds user's board to database
// It returns board's assigned ID and nil on success, any number and error on failure
func (boardApp *BoardApp) AddBoard(board *entity.Board) (int, error) {
	grpcBoard := grpcPins.Board{}
	ConvertToGrpcBoard(&grpcBoard, board)
	grpcBoardID, err := boardApp.grpcClient.AddBoard(context.Background(), &grpcBoard)
	if err != nil {
		if strings.Contains(err.Error(), entity.CreateBoardError.Error()) {
			return -1, entity.CreateBoardError
		}
		return -1, err
	}
	return int(grpcBoardID.BoardID), nil
}

// GetBoard returns board with passed boardID
// It returns that board and nil on success, nil and error on failure
func (boardApp *BoardApp) GetBoard(boardID int) (*entity.BoardInfo, error) {
	board, err := boardApp.grpcClient.GetBoard(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.BoardNotFoundError.Error()) {
			return nil, entity.BoardNotFoundError
		}
		return nil, err
	}
	boardInfo := &entity.BoardInfo{
		BoardID:     int(board.BoardID),
		UserID:      int(board.UserID),
		Title:       board.Title,
		Description: board.Description,
		ImageLink:   board.ImageLInk}
	return boardInfo, nil
}

// GetBoards returns all the boards with passed authorsID
// It returns slice of boards and nil on success, nil and error on failure
func (boardApp *BoardApp) GetBoards(authorID int) ([]entity.Board, error) {
	grpcBoardsList, err := boardApp.grpcClient.GetBoards(context.Background(), &grpcPins.UserID{Uid: int64(authorID)})
	if err != nil {
		return nil, err
	}
	return ConvertGrpcBoards(grpcBoardsList), nil
}

// DeleteBoard deletes user's board with passed boardID
// It returns nil on success and error on failure
func (boardApp *BoardApp) DeleteBoard(boardID int, userID int) error {
	initBoardID, err := boardApp.GetInitUserBoard(userID)
	if err != nil {
		return err
	}

	if boardID == initBoardID {
		return entity.DeleteInitBoardError
	}

	err = boardApp.CheckBoard(userID, boardID)
	if err != nil {
		return err
	}

	_, err = boardApp.grpcClient.DeleteBoard(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.DeleteBoardError.Error()) {
			return entity.DeleteBoardError
		}
		return err
	}

	return nil
}

func (boardApp *BoardApp) GetInitUserBoard(userID int) (int, error) {
	grpcBoardID, err := boardApp.grpcClient.GetInitUserBoard(context.Background(), &grpcPins.UserID{Uid: int64(userID)})
	if err != nil {
		return 0, err
	}
	return int(grpcBoardID.BoardID), nil
}

func (boardApp *BoardApp) CheckBoard(userID int, boardID int) error {
	board, err := boardApp.GetBoard(boardID)
	if err != nil {
		return err
	}

	if board.UserID != userID {
		return entity.CheckBoardOwnerError
	}
	return nil
}

func (boardApp *BoardApp) UploadBoardAvatar(boardID int, imageLink string) error {
	_, err := boardApp.grpcClient.UploadBoardAvatar(context.Background(), &grpcPins.FileInfo{
		BoardID: int64(boardID), ImagePath: imageLink,
	})
	if err != nil {
		if strings.Contains(err.Error(), entity.BoardAvatarUploadError.Error()) {
			return entity.BoardAvatarUploadError
		}
		return err
	}

	return nil
}

func ConvertToGrpcBoard(grpcPin *grpcPins.Board, pin *entity.Board) {
	grpcPin.UserID = int64(pin.UserID)
	grpcPin.BoardID = int64(pin.BoardID)
	grpcPin.Title = pin.Title
	grpcPin.Description = pin.Description
	grpcPin.ImageLInk = pin.ImageLInk
}

func ConvertFromGrpcBoard(board *entity.Board, grpcBoard *grpcPins.Board) {
	board.UserID = int(grpcBoard.UserID)
	board.BoardID = int(grpcBoard.BoardID)
	board.Title = grpcBoard.Title
	board.Description = grpcBoard.Description
	board.ImageLInk = grpcBoard.ImageLInk
}

func ConvertGrpcBoards(grpcBoards *grpcPins.BoardsList) []entity.Board {
	boards := make([]entity.Board, 0)
	for _, grpcBoard := range grpcBoards.Boards {
		board := entity.Board{}
		ConvertFromGrpcBoard(&board, grpcBoard)
		boards = append(boards, board)
	}
	return boards
}
