package pins

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Result struct {
	Body interface{} `json:"body,omitempty"`
	Err  string      `json:"err,omitempty"`
}

type pin struct {
	boardID     int
	pinId       int
	title       string
	description string
	imageLink   string
}
type userPinSet struct {
	userPins map[int][]pin
	userId   int
	pinId    int
	mutex    sync.RWMutex
}

type pinResult struct {
	BoardID     int    `json:"board"`
	PinId       int    `json:"pin_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageLink   string `json:"image_link"`
}

var MyPins userPinSet

func (pinSet *userPinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	pinInput := new(pinResult)
	err := decoder.Decode(pinInput)
	if err != nil {
		log.Printf("error while unmarshalling JSON: %s", err)
		w.Write([]byte(`{"code": 400}`))
		return
	}

	pinSet.mutex.Lock()

	pinSet.userPins[pinSet.userId] = append(pinSet.userPins[pinSet.userId], pin{
		boardID:     pinInput.BoardID,
		pinId:       pinInput.PinId,
		title:       pinInput.Title,
		description: pinInput.Description,
		imageLink:   pinInput.ImageLink,
	})

	pinSet.mutex.Unlock()

	w.Write([]byte(`{"code": 200}`)) // returning success code
}

func (pinSet *userPinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"bad pinId"}`, 400)
		return
	}
	pinsSet, err := pinSet.getPins()
	if err != nil {
		http.Error(w, `{"error":"db"}`, 500)
		return
	}

	for _, p := range pinsSet {
		if p.pinId == pinId {
			p = pinsSet[len(pinsSet)-1]
			pinsSet = pinsSet[:len(pinsSet)-1]
			break
		}
	}
}

func (pinSet *userPinSet) GetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"bad pinId"}`, 400)
		return
	}
	pinsSet, err := pinSet.getPins()
	if err != nil {
		http.Error(w, `{"error":"db"}`, 500)
		return
	}
	var resultPin pin
	for _, p := range pinsSet {
		if p.pinId == pinId {
			resultPin = p
			break
		}
	}
	body := map[string]interface{}{
		"pin": resultPin,
	}
	json.NewEncoder(w).Encode(&Result{Body: body})
}

func (pinSet *userPinSet) getPins() ([]pin, error) {
	pinSet.mutex.RLock()
	defer pinSet.mutex.RUnlock()

	return pinSet.userPins[pinSet.userId], nil
}
