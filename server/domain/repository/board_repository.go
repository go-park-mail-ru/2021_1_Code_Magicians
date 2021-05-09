package repository

import "pinterest/domain/entity"

type BoardRepository interface {
	AddBoard(board *entity.Board) (int, error)    // Creating user's board
	GetBoard(boardID int) (*entity.Board, error)  // Get description of the board
	GetBoards(userID int) ([]entity.Board, error) // Get boards by authorID
	DeleteBoard(boardID int) error                // Removes user's board by ID
	GetInitUserBoard(userID int) (int, error)     // Get initial user's board
	UploadBoardAvatar(boardID int, avatarLink string) error
}
