package pins

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	//"log"
	"net/http"
	"strconv"
	"sync"
)

func NewPinsSet(idUser int) *UserPinSet {
	return &UserPinSet{
		mutex:    sync.RWMutex{},
		userPins: map[int][]*Pin{},
		userId: idUser,
	}
}

type Result struct {
	Body interface{} `json:"body,omitempty"`
	Err  string      `json:"err,omitempty"`
}

/*type pin struct {
	boardID     int
	pinId       int
	title       string
	description string
	imageLink   string
}*/
type UserPinSet struct {
	userPins map[int][]*Pin
	userId   int
	pinId    int
	mutex    sync.RWMutex
}

type Pin struct {
	PinId       int    `json:"pin_id"`
	BoardID     int    `json:"board"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageLink   string `json:"image_link"`
}


func (pinSet *UserPinSet) PinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println(r.URL.Path)
	switch r.URL.Path {

	case "/pins/{id:[0-9]+}":
		if r.Method == http.MethodGet {
			pinSet.GetPinByID(w, r)
		} else if r.Method == http.MethodDelete {
			pinSet.DelPinByID(w, r)
		} else {
			w.Write([]byte(`{"code": 400}`))
			return
		}
	case "/pin/":
		if r.Method != http.MethodPost {
			w.Write([]byte(`{"code": 400}`))
			return
		}
		pinSet.AddPin(w, r)
	default:
		w.Write([]byte(`{"code": 400}`))
		return
	}
}

func (pinSet *UserPinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	pinId, _ := strconv.Atoi(r.FormValue("pin_id"))
	board, _ := strconv.Atoi(r.FormValue("board"))
	title := r.FormValue("title")
	description := r.FormValue("description")
	imageLink := r.FormValue("image_link")

	pinInput := &Pin{
		PinId: pinId,
		BoardID:     board,
		Title:       title,
		Description: description,
		ImageLink:   imageLink,
	}

	id := pinSet.pinId
	pinSet.pinId++

	pinSet.mutex.Lock()

	pinSet.userPins[pinSet.userId] = append(pinSet.userPins[pinSet.userId], pinInput)

	pinSet.mutex.Unlock()
	body := map[string]interface{}{
		"pid_id": id,
	}
	json.NewEncoder(w).Encode(&Result{Body: body})

	//w.Write([]byte(`{"code": 200}`)) // returning success code
}

func (pinSet *UserPinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
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
		if p.PinId == pinId {
			p = pinsSet[len(pinsSet)-1]
			pinsSet = pinsSet[:len(pinsSet)-1]
			break
		}
	}
}

func (pinSet *UserPinSet) GetPinByID(w http.ResponseWriter, r *http.Request) {
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
	var resultPin *Pin
	for _, p := range pinsSet {
		if p.PinId == pinId {
			resultPin = p
			break
		}
	}
	body := map[string]interface{}{
		"pin": resultPin,
	}
	json.NewEncoder(w).Encode(&Result{Body: body})
}

func (pinSet *UserPinSet) getPins() ([]*Pin, error) {
	pinSet.mutex.RLock()
	defer pinSet.mutex.RUnlock()

	return pinSet.userPins[pinSet.userId], nil
}
