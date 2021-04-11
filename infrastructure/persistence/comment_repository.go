package persistence

import (
	"github.com/jackc/pgx/v4"
	"pinterest/domain/entity"
)

type CommentsRepo struct {
	db *pgx.Conn
}

func (r *CommentsRepo)AddComment(*entity.Comment) error {
	return nil
}

func (r *CommentsRepo)DeleteComment(*entity.Comment) error  {
	return nil
}

func (r *CommentsRepo)EditComment(*entity.Comment) error  {
	return nil
}
