package pin

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
)

type PinInfo struct {
	pinApp application.PinAppInterface
	s3App  application.S3AppInterface
}

func NewPinInfo(pinApp application.PinAppInterface, s3App application.S3AppInterface) *PinInfo {
	return &PinInfo{
		pinApp: pinApp,
		s3App:  s3App,
	}
}

func (pinInfo *PinInfo) HandleAddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currPin := entity.Pin{}

	err = json.Unmarshal(data, &currPin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultPin := &entity.Pin{
		Title:       currPin.Title,
		Description: currPin.Description,
		ImageLink:   currPin.ImageLink,
	}

	userId := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	resultPin.PinId, err = pinInfo.pinApp.AddPin(userId, resultPin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"pin_id": ` + strconv.Itoa(resultPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(body))
}

func (pinInfo *PinInfo) HandleDelPinByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	err = pinInfo.pinApp.DeletePin(pinId, userId, pinInfo.s3App)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (pinInfo *PinInfo) HandleGetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pinId, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultPin, err := pinInfo.pinApp.GetPin(pinId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(resultPin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (pinInfo *PinInfo) HandleGetPinsByBoardID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars[string(entity.IDKey)])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardPins, err := pinInfo.pinApp.GetPins(boardId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(boardPins)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

const maxPostPictureBodySize = 8 * 1024 * 1024 // 8 mB
// HandleUploadPicture takes picture from request and assigns it to current pin
func (pinInfo *PinInfo) HandleUploadPicture(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodySize := r.ContentLength
	if bodySize < 0 { // No picture was passed
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if bodySize > int64(maxPostPictureBodySize) { // Picture is too large
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(bodySize)
	file, _, err := r.FormFile("pinImage")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = pinInfo.pinApp.UploadPicture(userID, file, pinInfo.s3App)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
