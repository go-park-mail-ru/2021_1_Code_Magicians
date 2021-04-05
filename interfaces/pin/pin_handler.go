package pin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"pinterest/domain/entity"

	"github.com/gorilla/mux"
)

func (pinSet *entity.PinSet) HandleAddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currPin := entity.Pin{
		PinId: pinSet.PinId,
	}

	err = json.Unmarshal(data, &currPin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := pinSet.PinId
	pinSet.PinId++

	pinInput := &entity.Pin{
		PinId:       id,
		BoardID:     currPin.BoardID,
		Title:       currPin.Title,
		Description: currPin.Description,
		ImageLink:   currPin.ImageLink,
	}

	pinSet.Mutex.Lock()

	pinSet.UserId = r.Context().Value("userID").(int)

	pinSet.UserPins[pinSet.UserId] = append(pinSet.UserPins[pinSet.UserId], pinInput)
	pinCopy := *pinInput
	pinSet.AllPins = append(pinSet.AllPins, &pinCopy)

	pinSet.Mutex.Unlock()

	body := `{"pin_id": ` + strconv.Itoa(currPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (pinSet *entity.PinSet) HandleDelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pinSet.UserId = r.Context().Value("userID").(int)

	pinsSet, err := pinSet.GetPins()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pinSet.Mutex.Lock()

	for _, p := range pinSet.AllPins {
		if p.PinId == pinId {
			*p = *pinSet.AllPins[len(pinsSet)-1]
			pinSet.AllPins = pinSet.AllPins[:len(pinsSet)-1]
			break
		}
	}

	for _, p := range pinSet.UserPins[pinSet.UserId] {
		if p.PinId == pinId {
			*p = *pinSet.UserPins[pinSet.UserId][len(pinsSet)-1]
			pinSet.UserPins[pinSet.UserId] = pinSet.UserPins[pinSet.UserId][:len(pinsSet)-1]
			w.WriteHeader(http.StatusNoContent)
			pinSet.Mutex.Unlock()
			return
		}
	}

	pinSet.Mutex.Unlock()

	w.WriteHeader(http.StatusNotFound)
}

func (pinSet *entity.PinSet) HandleGetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var resultPin *entity.Pin

	for _, p := range pinSet.AllPins {
		if p.PinId == pinId {
			resultPin = p
			break
		}
	}

	if resultPin == nil {
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

func (pinSet *entity.PinSet) HandleGetPinsByBoardID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardPins := make([]*entity.Pin, 0, 0)

	pinSet.Mutex.RLock()

	for _, pin := range pinSet.AllPins {
		if pin.BoardID == boardId {
			inputPin := *pin
			boardPins = append(boardPins, &inputPin)

		}
	}

	pinSet.Mutex.RUnlock()

	body, err := json.Marshal(boardPins)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (pinSet *entity.PinSet) getPins() ([]*entity.Pin, error) {
	pinSet.Mutex.RLock()
	defer pinSet.Mutex.RUnlock()

	return pinSet.UserPins[pinSet.UserId], nil
}
