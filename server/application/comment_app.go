package application

import (
	"context"
	"pinterest/domain/entity"
	grpcComments "pinterest/services/comments/proto"
	"strings"
)

type CommentApp struct {
	grpcClient grpcComments.CommentsClient
}

func NewCommentApp(grpcClient grpcComments.CommentsClient) *CommentApp {
	return &CommentApp{grpcClient: grpcClient}
}

type CommentAppInterface interface {
	AddComment(comment *entity.Comment) error        // Add comment to pin
	GetComments(pinID int) ([]entity.Comment, error) // Get pin's comments
	DeleteComment(comment *entity.Comment) error     // Delete pin's comment
	EditComment(comment *entity.Comment) error       // Edit pin's comment
}

func (commentApp *CommentApp) AddComment(comment *entity.Comment) error {
	grpcComment := grpcComments.Comment{
		PinComment: comment.PinComment,
		PinID:      int64(comment.PinID),
		UserID:     int64(comment.UserID),
	}
	FillGrpcComment(&grpcComment, comment)
	_, err := commentApp.grpcClient.AddComment(context.Background(), &grpcComment)
	if err != nil {
		if strings.Contains(err.Error(), entity.AddCommentError.Error()) {
			return entity.AddCommentError
		}
		return err
	}

	return nil
}

func (commentApp *CommentApp) GetComments(pinID int) ([]entity.Comment, error) {
	comments, err := commentApp.grpcClient.GetComments(context.Background(), &grpcComments.PinID{PinID: int64(pinID)})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.CommentsNotFoundError.Error()):
			return nil, entity.CommentsNotFoundError
		case strings.Contains(err.Error(), entity.GetCommentsError.Error()):
			return nil, entity.GetCommentsError
		case strings.Contains(err.Error(), entity.CommentScanError.Error()):
			return nil, entity.CommentScanError
		default:
			return nil, err
		}
	}
	resComments := ConvertGrpcComments(comments)

	return resComments, nil
}

func (commentApp *CommentApp) DeleteComment(comment *entity.Comment) error {
	return nil
}

func (commentApp *CommentApp) EditComment(comment *entity.Comment) error {
	return nil
}

func ConvertGrpcComments(grpcComments *grpcComments.CommentsList) []entity.Comment {
	comments := make([]entity.Comment, 0)
	for _, grpcComment := range grpcComments.Comments {
		comment := entity.Comment{}
		FillGrpcComment(grpcComment, &comment)
		comments = append(comments, comment)
	}
	return comments
}

func FillGrpcComment(grpcComment *grpcComments.Comment, comment *entity.Comment) {
	comment.PinID = int(grpcComment.PinID)
	comment.UserID = int(grpcComment.UserID)
	comment.PinComment = grpcComment.PinComment
}
