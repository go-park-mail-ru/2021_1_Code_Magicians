package usecase

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type CommentApp struct {
	c      repository.CommentRepository
	pinApp PinAppInterface
}

func NewCommentApp(c repository.CommentRepository, pinApp PinAppInterface) *CommentApp {
	return &CommentApp{
		c:      c,
		pinApp: pinApp,
	}
}

type CommentAppInterface interface {
	AddComment(comment *entity.Comment) error        // Add comment to pin
	GetComments(pinID int) ([]entity.Comment, error) // Get pin's comments
	DeleteComment(comment *entity.Comment) error     // Delete pin's comment
	EditComment(comment *entity.Comment) error       // Edit pin's comment
}

func (commentApp *CommentApp) AddComment(comment *entity.Comment) error {
	_, err := commentApp.pinApp.GetPin(comment.PinID)
	if err != nil {
		return err
	}

	return commentApp.c.AddComment(comment)
}

func (commentApp *CommentApp) GetComments(pinID int) ([]entity.Comment, error) {
	_, err := commentApp.pinApp.GetPin(pinID)
	if err != nil {
		return nil, err
	}

	return commentApp.c.GetComments(pinID)
}

func (commentApp *CommentApp) DeleteComment(comment *entity.Comment) error {
	return nil
}

func (commentApp *CommentApp) EditComment(comment *entity.Comment) error {
	return nil
}
