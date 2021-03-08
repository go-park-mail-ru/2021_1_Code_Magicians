package pins

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
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

type Result struct {
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
	PinId       int    `json:"pin_id"`
	BoardID     int    `json:"board"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageLink   string `json:"image_link"`
}

func (pinSet *UserPinSet) PinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//fmt.Println(r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		pinSet.GetPinByID(w, r)

	case http.MethodDelete:
		pinSet.DelPinByID(w, r)

	case http.MethodPost:
		pinSet.AddPin(w, r)

	default:
		http.Error(w, `{"error":"bad request"}`, 400)
		return
	}
}

func (pinSet *UserPinSet) AddPin(w http.ResponseWriter, r *http.Request) {
	board, _ := strconv.Atoi(r.FormValue("board"))
	title := r.FormValue("title")
	description := r.FormValue("description")
	imageLink := r.FormValue("image_link")

	id := pinSet.pinId
	pinSet.pinId++

	pinInput := &Pin{
		PinId:       id,
		BoardID:     board,
		Title:       title,
		Description: description,
		ImageLink:   imageLink,
	}

	pinSet.mutex.Lock()

	pinSet.userPins[pinSet.userId] = append(pinSet.userPins[pinSet.userId], pinInput)

	pinSet.mutex.Unlock()
	body := map[string]interface{}{
		"pin_id": id,
	}
	json.NewEncoder(w).Encode(&Result{Body: body})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code": 200}`)) // returning success code
}

func (pinSet *UserPinSet) DelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"bad pin_id"}`, 400)
		return
	}

	for _, p := range pinSet.userPins[pinSet.userId] {
		if p.PinId == pinId {
			fmt.Println(p.PinId, " ", (*p).PinId)
			p = pinSet.userPins[pinSet.userId][len(pinSet.userPins[pinSet.userId])-1]
			pinSet.userPins[pinSet.userId] = pinSet.userPins[pinSet.userId][:len(pinSet.userPins[pinSet.userId])-1]
			break
		}
		w.WriteHeader(http.StatusOK)
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
	if resultPin == nil {
		http.Error(w, `{"body":{"pin":null}}`, 404)
		return
	}
	body := map[string]interface{}{
		"pin": resultPin,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&Result{Body: body})
}

func (pinSet *UserPinSet) getPins() ([]*Pin, error) {
	pinSet.mutex.RLock()
	defer pinSet.mutex.RUnlock()

	return pinSet.userPins[pinSet.userId], nil
}
