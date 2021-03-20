package pins

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/auth"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type PinsStorage struct {
	Storage *UserPinSet
}

func NewPinsSet(idUser int) *UserPinSet {
	return &UserPinSet{
		mutex:    sync.RWMutex{},
		userPins: map[int][]*Pin{},
		userId:   idUser,
	}
}

type UserPinSet struct {
	userPins map[int][]*Pin
	userId   int
	pinId    int
	mutex    sync.RWMutex
}

type Pin struct {
	PinId       int    `json:"id"`
	BoardID     int    `json:"boardID"`
	Title       string `json:"title"`
	ImageLink   string `json:"pinImage"`
	Description string `json:"description"`
}

func (pinSet *UserPinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, found := auth.CheckCookies(r)
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currPin := Pin{
		PinId: pinSet.pinId,
	}

	err = json.Unmarshal(data, &currPin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := pinSet.pinId
	pinSet.pinId++

	pinInput := &Pin{
		PinId:       id,
		BoardID:     currPin.BoardID,
		Title:       currPin.Title,
		Description: currPin.Description,
		ImageLink:   currPin.ImageLink,
	}

	pinSet.mutex.Lock()

	pinSet.userPins[pinSet.userId] = append(pinSet.userPins[pinSet.userId], pinInput)

	pinSet.mutex.Unlock()

	body := `{"pin_id": ` + strconv.Itoa(currPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (pinSet *UserPinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
	_, found := auth.CheckCookies(r)
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pinsSet, err := pinSet.getPins()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, p := range pinSet.userPins[pinSet.userId] {
		if p.PinId == pinId {
			*p = *pinSet.userPins[pinSet.userId][len(pinsSet)-1]
			pinSet.userPins[pinSet.userId] = pinSet.userPins[pinSet.userId][:len(pinsSet)-1]
			w.WriteHeader(http.StatusOK)
			break
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (pinSet *UserPinSet) GetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pinsSet, err := pinSet.getPins()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resultPin *Pin

	for _, p := range pinsSet {
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

	_, err = w.Write(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (pinSet *UserPinSet) getPins() ([]*Pin, error) {
	pinSet.mutex.RLock()
	defer pinSet.mutex.RUnlock()

	return pinSet.userPins[pinSet.userId], nil
}
