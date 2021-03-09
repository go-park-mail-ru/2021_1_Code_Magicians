package pins

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

func NewPinsSet(idUser int) *UserPinSet {
	return &UserPinSet{
		mutex:    sync.RWMutex{},
		userPins: map[int][]*Pin{},
		userId:   idUser,
	}
}

type ResponseServer struct {
	Body interface{} `json:"body,omitempty"`
	Err  string      `json:"err,omitempty"`
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
	ImageLink   string `json:"imageLink"`
	Description string `json:"description"`
}

func (pinSet *UserPinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	currPin := Pin{
		PinId: pinSet.pinId,
	}

	err = json.Unmarshal(data, &currPin)
	fmt.Println(currPin)
	if err != nil {
		fmt.Println(err)
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

	w.WriteHeader(http.StatusCreated) // returning success code
	body := "{pin_id: " + strconv.Itoa(currPin.PinId) + "}"
	w.Write([]byte(body))
}

func (pinSet *UserPinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
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
			p = pinSet.userPins[pinSet.userId][len(pinsSet)-1]
			pinSet.userPins[pinSet.userId] = pinSet.userPins[pinSet.userId][:len(pinsSet)-1]
			break
		}
	}
	w.WriteHeader(http.StatusOK)
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

	body := map[string]interface{}{
		"pin": resultPin,
	}

	err = json.NewEncoder(w).Encode(&ResponseServer{Body: body})
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
