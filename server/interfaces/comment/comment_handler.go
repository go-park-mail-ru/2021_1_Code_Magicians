package comment

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/domain/entity"
	"strconv"

	"go.uber.org/zap"

	"pinterest/application"

	"github.com/gorilla/mux"
)

type CommentInfo struct {
	commentApp application.CommentAppInterface
	pinApp     application.PinAppInterface
	logger     *zap.Logger
}

func NewCommentInfo(commentApp application.CommentAppInterface,
	pinApp application.PinAppInterface,
	logger *zap.Logger) *CommentInfo {
	return &CommentInfo{
		commentApp: commentApp,
		pinApp:     pinApp,
		logger:     logger,
	}
}

func (commentInfo *CommentInfo) HandleAddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currComment := entity.Comment{}

	err = json.Unmarshal(data, &currComment)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultComment := &entity.Comment{
		UserID:     userID,
		PinID:      pinID,
		PinComment: currComment.PinComment,
	}

	err = commentInfo.commentApp.AddComment(resultComment)
	if err != nil {
		commentInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		switch err {
		case entity.PinNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	comment := entity.CommentTextOutput{Text: currComment.PinComment}
	body, err := json.Marshal(comment)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (commentInfo *CommentInfo) HandleGetComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pinComments, err := commentInfo.commentApp.GetComments(pinID)
	if err != nil && err != entity.CommentsNotFoundError {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	comments := entity.CommentsOutput{Comments: pinComments}
	if comments.Comments == nil {
		comments.Comments = make([]entity.Comment, 0)
	}

	body, err := json.Marshal(comments)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
