package comment

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
)

type CommentInfo struct {
	commentApp application.CommentAppInterface
	pinApp     application.PinAppInterface
	logger *zap.Logger
}

func NewCommentInfo(commentApp application.CommentAppInterface,
	pinApp application.PinAppInterface,
	logger *zap.Logger) *CommentInfo {
	return &CommentInfo{
		commentApp: commentApp,
		pinApp:     pinApp,
		logger: logger,
	}
}

func (commentInfo *CommentInfo) HandleAddComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	pinID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	_, err = commentInfo.pinApp.GetPin(pinID)
	if err != nil {
		commentInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	comment := entity.CommentTextOutput{currComment.PinComment}
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
	pinId, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = commentInfo.pinApp.GetPin(pinId)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pinComments, err := commentInfo.commentApp.GetComments(pinId)
	if err != nil {
		commentInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comments := entity.CommentsOutput{pinComments}


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
