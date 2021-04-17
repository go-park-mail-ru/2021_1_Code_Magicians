package persistence

import (
	"context"
	"errors"
	"fmt"
	"pinterest/domain/entity"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type BoardsRepo struct {
	db *pgxpool.Pool
}

func NewBoardsRepository(db *pgxpool.Pool) *BoardsRepo {
	return &BoardsRepo{db}
}

const createBoardQuery string = "INSERT INTO Boards (userID, title, description)\n" +
	"values ($1, $2, $3)\n" +
	"RETURNING boardID"

// AddBoard add new board to database with passed fields
// It returns board's assigned ID and nil on success, any number and error on failure
func (r *BoardsRepo) AddBoard(board *entity.Board) (int, error) {
	row := r.db.QueryRow(context.Background(), createBoardQuery, board.UserID, board.Title, board.Description)
	newBoardID := 0
	err := row.Scan(&newBoardID)
	if err != nil {
		// Other errors
		return -1, err
	}
	return newBoardID, nil
}

const deleteBoardQuery string = "DELETE FROM Boards WHERE boardID=$1 AND userID=$2"

// DeleteBoard deletes board with passed id belonging to passed user.
// It returns error if board is not found or if there were problems with database
func (r *BoardsRepo) DeleteBoard(boardID int, userID int) error {
	commandTag, err := r.db.Exec(context.Background(), deleteBoardQuery, boardID, userID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("Board not found")
	}
	return err
}

const getBoardQuery string = "SELECT userID, title, description, imageLink FROM Boards WHERE boardID=$1"

// GetBoard fetches board with passed ID from database
// It returns that board, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoard(boardID int) (*entity.Board, error) {
	board := entity.Board{BoardID: boardID}
	row := r.db.QueryRow(context.Background(), getBoardQuery, boardID)
	err := row.Scan(&board.UserID, &board.Title, &board.Description, &board.ImageLInk)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No board found with such id")
		}

		// Other errors
		return nil, err
	}
	return &board, nil
}

const getBoardsByUserQuery string = "SELECT boardID, title, description, imageLink FROM Boards WHERE userID=$1"

// GetBoards fetches all boards created by user with specified ID from database
// It returns slice of these boards, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoards(userID int) ([]entity.Board, error) {
	boards := make([]entity.Board, 0)
	rows, err := r.db.Query(context.Background(), getBoardsByUserQuery, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("No boards found in database with passed userID")
		}
		return nil, err
	}

	for rows.Next() {
		board := entity.Board{UserID: userID}
		err = rows.Scan(&board.BoardID, &board.Title, &board.Description, &board.ImageLInk)
		if err != nil {
			return nil, err // TODO: error handling
		}
		boards = append(boards, board)
	}
	return boards, nil
}

const getInitUserBoardQuery string = "SELECT b1.boardID, b1.title, b1.description, b1.imageLink\n" +
	"FROM boards AS b1\n" +
	"INNER JOIN boards AS b2 on b2.boardID = b1.boardID AND b2.userID = $1\n" +
	"GROUP BY b1.boardID, b2.userID\n" +
	"ORDER BY b2.userID LIMIT 1;"

// GetInitUserBoard gets user's first board from database
// It returns that board and nil on success, nil and error on failure
func (r *BoardsRepo) GetInitUserBoard(userID int) (int, error) {
	board := entity.Board{UserID: userID}
	row := r.db.QueryRow(context.Background(), getInitUserBoardQuery, userID)
	err := row.Scan(&board.BoardID, &board.Title, &board.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, fmt.Errorf("No board found")
		}
		return -1, err
	}
	return board.BoardID, nil
}

const checkBoardQuery string = "SELECT boardID\n" +
	"FROM boards\n" +
	"WHERE boardID = $1 AND userID = $2;"

// CheckBoard checking that passed board belongs to passed user
// It returns that nil on success, error on failure
func (r *BoardsRepo) CheckBoard(userID int, boardID int) error {
	id := 0
	row := r.db.QueryRow(context.Background(), checkBoardQuery, boardID, userID)
	err := row.Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("That board is not associated with that user")
		}
		return err
	}
	return nil
}

const saveBoardPictureQuery string = "UPDATE boards\n" +
	"SET imageLink=$1\n" +
	"WHERE boardID=$2"

func (r *BoardsRepo) UploadBoardAvatar(boardID int, imageLink string) error {
	commandTag, err := r.db.Exec(context.Background(), saveBoardPictureQuery, imageLink, boardID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("Board not found")
	}
	return nil
}
