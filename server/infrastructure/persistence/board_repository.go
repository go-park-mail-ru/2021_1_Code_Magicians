package persistence

import (
	"context"
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
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	row := tx.QueryRow(context.Background(), createBoardQuery, board.UserID, board.Title, board.Description)
	newBoardID := 0
	err = row.Scan(&newBoardID)
	if err != nil {
		return -1, entity.CreateBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return newBoardID, nil
}

const deleteBoardQuery string = "DELETE FROM Boards WHERE boardID=$1"

// DeleteBoard deletes board with passed id belonging to passed user.
// It returns error if board is not found or if there were problems with database
func (r *BoardsRepo) DeleteBoard(boardID int) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), deleteBoardQuery, boardID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.DeleteBoardError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return err
}

const getBoardQuery string = "SELECT userID, title, description FROM Boards WHERE boardID=$1"

// GetBoard fetches board with passed ID from database
// It returns that board, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoard(boardID int) (*entity.Board, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	board := entity.Board{BoardID: boardID}
	row := tx.QueryRow(context.Background(), getBoardQuery, boardID)
	err = row.Scan(&board.UserID, &board.Title, &board.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.BoardNotFoundError
		}

		// Other errors
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return &board, nil
}

const getBoardsByUserQuery string = "SELECT boardID, title, description FROM Boards WHERE userID=$1"

// GetBoards fetches all boards created by user with specified ID from database
// It returns slice of these boards, nil on success and nil, error on failure
func (r *BoardsRepo) GetBoards(userID int) ([]entity.Board, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	boards := make([]entity.Board, 0)
	rows, err := tx.Query(context.Background(), getBoardsByUserQuery, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.GetBoardsByUserIDError
		}
		return nil, err
	}

	for rows.Next() {
		board := entity.Board{UserID: userID}
		err = rows.Scan(&board.BoardID, &board.Title, &board.Description)
		if err != nil {
			return nil, err // TODO: error handling
		}
		boards = append(boards, board)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return boards, nil
}

const getInitUserBoardQuery string = "SELECT b1.boardID, b1.title, b1.description\n" +
	"FROM boards AS b1\n" +
	"INNER JOIN boards AS b2 on b2.boardID = b1.boardID AND b2.userID = $1\n" +
	"GROUP BY b1.boardID, b2.userID\n" +
	"ORDER BY b2.userID LIMIT 1;"

// GetInitUserBoard gets user's first board from database
// It returns that board and nil on success, nil and error on failure
func (r *BoardsRepo) GetInitUserBoard(userID int) (int, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return -1, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	board := entity.Board{UserID: userID}
	row := tx.QueryRow(context.Background(), getInitUserBoardQuery, userID)
	err = row.Scan(&board.BoardID, &board.Title, &board.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, entity.NotFoundInitUserBoard
		}
		return -1, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return -1, entity.TransactionCommitError
	}
	return board.BoardID, nil
}

const saveBoardPictureQuery string = "UPDATE boards\n" +
	"SET imageLink=$1\n" +
	"WHERE boardID=$2"

func (r *BoardsRepo) UploadBoardAvatar(boardID int, imageLink string) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(), saveBoardPictureQuery, imageLink, boardID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.FileUploadError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}