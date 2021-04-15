package pin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
)

type PinInfo struct {
	pinApp   application.PinAppInterface
	boardApp application.BoardAppInterface
	s3App    application.S3AppInterface
}

func NewPinInfo(pinApp application.PinAppInterface,
	s3App application.S3AppInterface,
	boardApp application.BoardAppInterface) *PinInfo {
	return &PinInfo{
		pinApp:   pinApp,
		boardApp: boardApp,
		s3App:    s3App,
	}
}

const maxPostPictureBodySize int = 8 * 1024 * 1024 // 8 mB
const maxJSONSize int = 1024 * 1024

func (pinInfo *PinInfo) HandleAddPin(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println(r.Header)
	fmt.Println(r.Body)
	r.ParseMultipartForm(bodySize)
	jsonData := r.FormValue("pinInfo") // TODO: replace string constants with keys
	currPin := entity.Pin{}
	fmt.Println(jsonData)
	err := json.Unmarshal([]byte(jsonData), &currPin)
	fmt.Println(jsonData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	currPin.UserID = userID
	fmt.Println("---------------------------------------------6")
	if currPin.BoardID != 0 {
		err = pinInfo.boardApp.CheckBoard(userID, currPin.BoardID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	fmt.Println("---------------------------------------------4")
	currPin.PinId, err = pinInfo.pinApp.CreatePin(&currPin)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("---------------------------------------------3")
	file, _, err := r.FormFile("pinImage")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("---------------------------------------------2")
	err = pinInfo.pinApp.UploadPicture(currPin.PinId, file)
	fmt.Println("---------------------------------------------1")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	// TODO: Add pin to specified board
}

func (pinInfo *PinInfo) HandleAddPinToBoard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	err = pinInfo.boardApp.CheckBoard(userID, boardID)
	fmt.Println(err)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pinID, err := strconv.Atoi(vars["pinID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = pinInfo.pinApp.AddPin(boardID, pinID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (pinInfo *PinInfo) HandleSavePin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	err = pinInfo.pinApp.SavePin(userId, pinId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (pinInfo *PinInfo) HandleDelPinByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = pinInfo.boardApp.CheckBoard(userID, boardID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pinID, err := strconv.Atoi(vars["pinID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = pinInfo.pinApp.DeletePin(boardID, pinID)
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

	pinsBody, err := json.Marshal(boardPins)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"pins": ` + string(pinsBody) + `}`

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(body))
}

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
	err = pinInfo.pinApp.UploadPicture(userID, file) // TODO: change userID to pinID

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
