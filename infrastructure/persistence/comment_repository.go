package persistence

import (
	"context"
	"errors"
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

func (r *CommentsRepo) AddComment(comment *entity.Comment) error {
	tx, err := r.db.Begin(context.Background())
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
		return errors.New("Error during posting the comment")
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return entity.TransactionCommitError
	}
	return nil
}

const getCommentsByPinQuery string = "SELECT userID, pinID, text FROM comments\n" +
	"WHERE pinID=$1;"

func (r *CommentsRepo) GetComments(pinID int) ([]entity.Comment, error) {
	tx, err := r.db.Begin(context.Background())
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
		return nil, err
	}

	for rows.Next() {
		comment := entity.Comment{}
		err := rows.Scan(&comment.UserID, &comment.PinID, &comment.PinComment)
		if err != nil {
			return nil, err // TODO: error handling
		}
		comments = append(comments, comment)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, entity.TransactionCommitError
	}
	return comments, nil
}

func (r *CommentsRepo) DeleteComment(*entity.Comment) error {
	tx, err := r.db.Begin(context.Background())
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

func (r *CommentsRepo) EditComment(*entity.Comment) error {
	tx, err := r.db.Begin(context.Background())
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
