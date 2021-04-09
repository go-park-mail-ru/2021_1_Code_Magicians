package pin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
)

type PinInfo struct {
	PinApp application.PinAppInterface
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
		BoardID:     currPin.BoardID,
		Title:       currPin.Title,
		Description: currPin.Description,
		ImageLink:   currPin.ImageLink,
	}

	resultPin.PinId, err = pinInfo.PinApp.AddPin(resultPin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"pin_id": ` + strconv.Itoa(resultPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (pinInfo *PinInfo) HandleDelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID

	err = pinInfo.PinApp.DeletePin(pinId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (pinInfo *PinInfo) HandleGetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultPin, err := pinInfo.PinApp.GetPin(pinId)
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
	w.Write(body)
}

func (pinInfo *PinInfo) HandleGetPinsByBoardID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardPins, err := pinInfo.PinApp.GetPins(boardId)
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
	w.Write(body)
}
