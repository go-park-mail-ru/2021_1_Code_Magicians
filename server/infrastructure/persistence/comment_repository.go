package persistence

import (
	"context"
	"pinterest/domain/entity"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CommentsRepo struct {
	db *pgxpool.Pool
}

func NewCommentsRepository(db *pgxpool.Pool) *CommentsRepo {
	return &CommentsRepo{db}
}

const addCommentQuery string = "INSERT INTO comments (userID, pinID, text)\n" +
	"values ($1, $2, $3);"

func (s *CommentsRepo) AddComment(comment *entity.Comment) error {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(),
		addCommentQuery,
		comment.UserID,
		comment.PinID,
		comment.PinComment)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return entity.AddCommentError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const getCommentsByPinQuery string = "SELECT userID, pinID, text FROM comments\n" +
	"WHERE pinID=$1;"

func (s *CommentsRepo) GetComments(pinID int) ([]entity.Comment, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	comments := make([]entity.Comment, 0)
	rows, err := tx.Query(context.Background(), getCommentsByPinQuery, pinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, entity.GetCommentsError
	}

	for rows.Next() {
		comment := entity.Comment{}
		err = rows.Scan(&comment.UserID, &comment.PinID, &comment.PinComment)
		if err != nil {
			return nil, entity.ReturnCommentsError
		}
		comments = append(comments, comment)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return comments, nil
}

func (s *CommentsRepo) DeleteComment(*entity.Comment) error {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

func (s *CommentsRepo) EditComment(*entity.Comment) error {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}
