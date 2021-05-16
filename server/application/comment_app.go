package application

import (
	"context"
	"pinterest/domain/entity"
	grpcComments "pinterest/services/comments/proto"
)

type CommentApp struct {
	grpcClient  grpcComments.CommentsClient
}

func NewCommentApp(grpcClient  grpcComments.CommentsClient) *CommentApp {
	return &CommentApp{grpcClient: grpcClient}
}

type CommentAppInterface interface {
	AddComment(*entity.Comment) error // Add comment to pin
	GetComments(int) ([]entity.Comment, error)
	DeleteComment(*entity.Comment) error // Delete pin's comment
	EditComment(*entity.Comment) error  // Edit pin's comment
}


func (com *CommentApp)AddComment(comment *entity.Comment) error {
	grpcComment:= grpcComments.Comment{
		PinComment: comment.PinComment,
		PinID: int64(comment.PinID),
		UserID: int64(comment.UserID),
	}
	FillGrpcComment(&grpcComment, comment)
	_, err := com.grpcClient.AddComment(context.Background(), &grpcComment)
	return err
}

func (com *CommentApp) GetComments(pinID int)  ([]entity.Comment, error) {
	comments, err := com.grpcClient.GetComments(context.Background(), &grpcComments.PinID{PinID: int64(pinID)})
	if err != nil {
		return nil, err
	}
	resComments := ConvertGrpcComments(comments)

	return resComments, nil
}

func (com *CommentApp)DeleteComment(comment *entity.Comment) error  {
	return nil
}

func (com *CommentApp)EditComment(comment *entity.Comment) error  {
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