package repository

import "pinterest/domain/entity"

type CommentRepository interface {
	AddComment(comment *entity.Comment) error // Add comment to pin
	GetComments(commentID int) ([]entity.Comment, error)
	DeleteComment(comment *entity.Comment) error // Delete pin's comment
	EditComment(comment *entity.Comment) error   // Edit pin's comment
}
