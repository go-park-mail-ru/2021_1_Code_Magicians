package persistence

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"pinterest/domain/entity"
)

type CommentsRepo struct {
	db *pgx.Conn
}

func NewCommentsRepository(db *pgx.Conn) *CommentsRepo {
	return &CommentsRepo{db}
}

const addCommentQuery string = "INSERT INTO comments (userID, pinID, text)\n" +
	"values ($1, $2, $3);"

func (r *CommentsRepo) AddComment(comment *entity.Comment) error {
	commandTag, err := r.db.Exec(context.Background(),
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
	return nil
}

const getCommentsByPinQuery string = "SELECT userID, pinID, text FROM comments\n" +
	"WHERE pinID=$1;"

func (r *CommentsRepo) GetComments(pinID int) ([]entity.Comment, error) {
	comments := make([]entity.Comment, 0)
	rows, err := r.db.Query(context.Background(), getCommentsByPinQuery, pinID)
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
	return comments, nil
}

func (r *CommentsRepo) DeleteComment(*entity.Comment) error {
	return nil
}

func (r *CommentsRepo) EditComment(*entity.Comment) error {
	return nil
}
