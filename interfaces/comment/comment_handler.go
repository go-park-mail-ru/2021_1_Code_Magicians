package comment

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"
)

type CommentInfo struct {
	CommentApp application.CookieApp
	PinApp application.PinAppInterface
	UserApp  application.UserApp
}

func (commentInfo *CommentInfo) HandleAddComment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currComment:= entity.Comment{}

	err = json.Unmarshal(data, &currComment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	resultComment := &entity.Comment{
		UserID:       userId,
		Description: currPin.Description,
		ImageLink:   currPin.ImageLink,
	}

	userId := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	resultPin.PinId, err = commentInfo.PinApp.AddPin(userId, resultPin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"pin_id": ` + strconv.Itoa(resultPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}
