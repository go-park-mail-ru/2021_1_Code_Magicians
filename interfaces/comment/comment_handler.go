package comment

import (
	"encoding/json"
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
}

func NewCommentInfo(commentApp application.CommentAppInterface, pinApp application.PinAppInterface) *CommentInfo {
	return &CommentInfo{
		commentApp: commentApp,
		pinApp:     pinApp,
	}
}

func (commentInfo *CommentInfo) HandleAddComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = commentInfo.pinApp.GetPin(pinId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currComment := entity.Comment{}

	err = json.Unmarshal(data, &currComment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	resultComment := &entity.Comment{
		UserID:     userId,
		PinID:      pinId,
		PinComment: currComment.PinComment,
	}

	err = commentInfo.commentApp.AddComment(resultComment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"text": "` + resultComment.PinComment + `"}`

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(body))
}

func (commentInfo *CommentInfo) HandleGetComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = commentInfo.pinApp.GetPin(pinId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pinComments, err := commentInfo.commentApp.GetComments(pinId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(pinComments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
