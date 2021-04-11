package application

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
	DeleteComment(*entity.Comment) error // Delete pin's comment
	EditComment(*entity.Comment) error  // Edit pin's comment
}
