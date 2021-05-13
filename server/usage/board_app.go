package usage

import (
	"context"
	"pinterest/domain/entity"
	grpcPins "pinterest/services/pins/proto"
)

type BoardApp struct {
	grpcClient grpcPins.PinsClient

}

func NewBoardApp(grpcClient grpcPins.PinsClient) *BoardApp {
	return &BoardApp{grpcClient}
}

type BoardAppInterface interface {
	AddBoard(*entity.Board) (int, error)     // Creating user's board
	GetBoard(int) (*entity.BoardInfo, error) // Get description of the board
	GetBoards(int) ([]entity.Board, error)   // Get boards by authorID
	GetInitUserBoard(int) (int, error)
	DeleteBoard(int, int) error // Removes user's board by ID
	CheckBoard(int, int) error
	UploadBoardAvatar(int, string) error
}

// AddBoard adds user's board to database
// It returns board's assigned ID and nil on success, any number and error on failure
func (brd *BoardApp) AddBoard(board *entity.Board) (int, error) {
	grpcBoard := grpcPins.Board{}
	ConvertToGrpcBoard(&grpcBoard, board)
	grpcBoardID, err := brd.grpcClient.AddBoard(context.Background(), &grpcBoard)
	if err != nil {
		return 0, err
	}
	return int(grpcBoardID.BoardID), nil
}

// GetBoard returns board with passed boardID
// It returns that board and nil on success, nil and error on failure
func (brd *BoardApp) GetBoard(boardID int) (*entity.BoardInfo, error) {
	board, err := brd.grpcClient.GetBoard(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	if err != nil {
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
func (brd *BoardApp) GetBoards(authorID int) ([]entity.Board, error) {
	grpcBoardsList, err := brd.grpcClient.GetBoards(context.Background(), &grpcPins.UserID{Uid: int64(authorID)})
	if err != nil {
		return nil, err
	}
	return ConvertGrpcBoards(grpcBoardsList), nil
}

// DeleteBoard deletes user's board with passed boardID
// It returns nil on success and error on failure
func (brd *BoardApp) DeleteBoard(boardID int, userID int) error {
	initBoardID, err := brd.GetInitUserBoard(userID)
	if err != nil {
		return err
	}

	if boardID == initBoardID {
		return entity.DeleteInitBoardError
	}

	err = brd.CheckBoard(userID, boardID)
	if err != nil {
		return err
	}

	_, err = brd.grpcClient.DeleteBoard(context.Background(), &grpcPins.BoardID{BoardID: int64(boardID)})
	return err
}

func (brd *BoardApp) GetInitUserBoard(userID int) (int, error) {
	grpcBoardID, err := brd.grpcClient.GetInitUserBoard(context.Background(), &grpcPins.UserID{Uid: int64(userID)})
	if err != nil {
		return 0, err
	}
	return int(grpcBoardID.BoardID), nil
}

func (brd *BoardApp) CheckBoard(userID int, boardID int) error {
	board, err := brd.GetBoard(boardID)
	if err != nil {
		return err
	}

	if board.UserID != userID {
		return entity.CheckBoardOwnerError
	}
	return nil
}

func (brd *BoardApp) UploadBoardAvatar(boardID int, imageLink string) error {
	_, err := brd.grpcClient.UploadBoardAvatar(context.Background(), &grpcPins.FileInfo{
		BoardID: int64(boardID), ImagePath: imageLink,
	})
	return err
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