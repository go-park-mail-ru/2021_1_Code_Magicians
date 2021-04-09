package persistence

import (
	"context"
	"fmt"
	"pinterest/domain/entity"

	"github.com/jackc/pgx/v4"
)

type BoardsRepo struct {
	db *pgx.Conn
}

func NewBoardsRepository(db *pgx.Conn) *BoardsRepo {
	return &BoardsRepo{db}
}

const createBoardQuery string = "INSERT INTO Boards (title, description)\n" +
	"values ($1, $2)\n" +
	"RETURNING boardID"

// BoardsRepo add new user to database with passed fields
// It returns user's assigned ID and nil on success, any number and error on failure
func (r *BoardsRepo) AddBoard(board *entity.Board) (int, error) {
	row := r.db.QueryRow(context.Background(), createBoardQuery, board.Title, board.Description)
	newBoardID := 0
	err := row.Scan(&newBoardID)
	if err != nil {
		// Other errors
		// log.Println(err)
		return -1, err
	}
	return newBoardID, nil
}

const deleteBoardQuery string = "DELETE FROM Boards WHERE boardID=$1 AND userID=$2"

// SaveUser deletes user with passed ID
// It returns nil on success and error on failure
func (r *BoardsRepo) DeleteBoard(boardID int, userID int) error {
	_, err := r.db.Exec(context.Background(), deleteBoardQuery, boardID, userID)
	return err
}

const getBoardQuery string = "SELECT userID, title, description FROM Boards WHERE boardID=$1"

// GetUser fetches user with passed ID from database
// It returns that user, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoard(boardID int) (*entity.Board, error) {
	board := entity.Board{BoardID: boardID}
	row := r.db.QueryRow(context.Background(), getBoardQuery, boardID)
	err := row.Scan(&board.UserID, &board.Title, &board.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No board found with such id")
		}
		// Other errors
		return nil, err
	}
	return &board, nil
}

const getBoardsByUserQuery string = "SELECT boardID, title, description FROM Boards WHERE userID=$1"

// GetUsers fetches all users from database
// It returns slice of all users, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoards(userID int) ([]entity.Board, error) {
	boards := make([]entity.Board, 0)
	rows, err := r.db.Query(context.Background(), getPinsByBoardQuery, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No boards found in database with passed userID")
		}

		// Other errors
		return nil, err
	}

	for rows.Next() {
		board := entity.Board{}
		err := rows.Scan(&board.BoardID, &board.UserID, &board.Title, &board.Description)
		if err != nil {
			return nil, err // TODO: error handling
		}
		boards = append(boards, board)
	}
	return boards, nil
}