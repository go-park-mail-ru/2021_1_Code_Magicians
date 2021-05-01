package repository

import "pinterest/domain/entity"

type CommentRepository interface {
	AddComment(*entity.Comment) error // Add comment to pin
	GetComments(int) ([]entity.Comment, error)
	DeleteComment(*entity.Comment) error // Delete pin's comment
	EditComment(*entity.Comment) error  // Edit pin's comment
}

