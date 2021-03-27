package boards

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)



type PinsStorage struct {
	Storage *PinSet
}

func NewPinsSet() *PinSet {
	return &PinSet{
		mutex:    sync.RWMutex{},
		userPins: map[int][]*Pin{},
		allPins: []*Pin{},
	}
}

type PinSet struct {
	userPins map[int][]*Pin
	allPins []*Pin
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

func (pinSet *PinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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

	pinSet.userId = r.Context().Value("userID").(int)

	pinSet.userPins[pinSet.userId] = append(pinSet.userPins[pinSet.userId], pinInput)
	pinCopy := *pinInput
	pinSet.allPins = append(pinSet.allPins, &pinCopy)

	pinSet.mutex.Unlock()

	body := `{"pin_id": ` + strconv.Itoa(currPin.PinId) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (pinSet *PinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if pinId > pinSet.allPins[len(pinSet.allPins) - 1].PinId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pinSet.userId = r.Context().Value("userID").(int)

	pinsSet, err := pinSet.getPins()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pinSet.mutex.Lock()

	for _, p := range pinSet.allPins {
		if p.PinId == pinId {
			*p = *pinSet.allPins[len(pinsSet)-1]
			pinSet.allPins = pinSet.allPins[:len(pinsSet)-1]
			break
		}
	}

	for _, p := range pinSet.userPins[pinSet.userId] {
		if p.PinId == pinId {
			*p = *pinSet.userPins[pinSet.userId][len(pinsSet)-1]
			pinSet.userPins[pinSet.userId] = pinSet.userPins[pinSet.userId][:len(pinsSet)-1]
			w.WriteHeader(http.StatusOK)
			pinSet.mutex.Unlock()
			return
		}
	}

	pinSet.mutex.Unlock()
	w.WriteHeader(http.StatusNotFound)
}

func (pinSet *PinSet) GetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if pinId > pinSet.allPins[len(pinSet.allPins) - 1].PinId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resultPin *Pin

	for _, p := range pinSet.allPins {
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

func (pinSet *PinSet) getPins() ([]*Pin, error) {
	pinSet.mutex.RLock()
	defer pinSet.mutex.RUnlock()

	return pinSet.userPins[pinSet.userId], nil
}
