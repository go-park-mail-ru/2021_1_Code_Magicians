package delivery

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type CommentApp struct {
	c repository.CommentRepository
}

func NewCommentApp(c repository.CommentRepository) *CommentApp {
	return &CommentApp{c}
}

type CommentAppInterface interface {
	AddComment(*entity.Comment) error // Add comment to pin
	GetComments(int) ([]entity.Comment, error)
	DeleteComment(*entity.Comment) error // Delete pin's comment
	EditComment(*entity.Comment) error   // Edit pin's comment
}

func (com *CommentApp) AddComment(comment *entity.Comment) error {
	return com.c.AddComment(comment)
}

func (com *CommentApp) GetComments(pinID int) ([]entity.Comment, error) {
	return com.c.GetComments(pinID)
}

func (com *CommentApp) DeleteComment(comment *entity.Comment) error {
	return nil
}

func (com *CommentApp) EditComment(comment *entity.Comment) error {
	return nil
}
