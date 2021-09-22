package comments

import (
	"context"
	"pinterest/domain/entity"
	. "pinterest/services/comments/proto"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *service {
	return &service{db}
}

const addCommentQuery string = "INSERT INTO comments (userID, pinID, text)\n" +
	"values ($1, $2, $3);"

func (s *service) AddComment(ctx context.Context, comment *Comment) (*Error, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	commandTag, err := tx.Exec(context.Background(),
		addCommentQuery,
		comment.UserID,
		comment.PinID,
		comment.PinComment)
	if err != nil {
		return &Error{}, err
	}
	if commandTag.RowsAffected() != 1 {
		return &Error{}, entity.AddCommentError
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &Error{}, entity.TransactionCommitError
	}
	return &Error{}, nil
}

const getCommentsByPinQuery string = "SELECT userID, pinID, text FROM comments\n" +
	"WHERE pinID=$1;"

func (s *service) GetComments(ctx context.Context, pinID *PinID) (*CommentsList, error) {
	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return &CommentsList{}, entity.TransactionBeginError
	}
	defer tx.Rollback(context.Background())

	comments := make([]*Comment, 0)
	rows, err := tx.Query(context.Background(), getCommentsByPinQuery, pinID.PinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &CommentsList{}, entity.CommentsNotFoundError
		}
		return &CommentsList{}, entity.GetCommentsError
	}

	for rows.Next() {
		comment := Comment{}
		err = rows.Scan(&comment.UserID, &comment.PinID, &comment.PinComment)
		if err != nil {
			return &CommentsList{}, entity.CommentScanError
		}
		comments = append(comments, &comment)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return &CommentsList{}, entity.TransactionCommitError
	}
	return &CommentsList{Comments: comments}, nil
}
