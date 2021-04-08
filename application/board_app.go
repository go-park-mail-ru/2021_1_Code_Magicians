package application

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type BoardApp struct {
	b repository.BoardRepository
}

func NewBoardApp(b repository.BoardRepository) *BoardApp {
	return &BoardApp{b}
}

type BoardRepository interface {
	AddBoard(board *entity.Board) (int, error) // Creating user's board
	GetBoard(int) (*entity.Board, error)       // Get description of the board
	GetBoards(int) ([]entity.Board, error)     // Get boards by authorID
	DeleteBoard(int) error                     // Removes user's board by ID
}

// AddBoard adds user's board to database
// It returns board's assigned ID and nil on success, any number and error on failure
func (brd *BoardApp) AddBoard(board *entity.Board) (int, error) {
	return brd.b.AddBoard(board)
}

// GetBoard returns board with passed boardID
// It returns that board and nil on success, nil and error on failure
func (brd *BoardApp) GetBoard(boardID int) (*entity.Board, error) {
	return brd.b.GetBoard(boardID)
}

// GetBoards returns all the boards with passed authorsID
// It returns slice of boards and nil on success, nil and error on failure
func (brd *BoardApp) GetBoards(authorID int) ([]entity.Board, error) {
	return brd.b.GetBoards(authorID)
}

// DeleteBoard deletes user's board with passed boardID
// It returns nil on success and error on failure
func (brd *BoardApp) DeleteBoard(boardID int) error {
	return brd.b.DeleteBoard(boardID)
}
