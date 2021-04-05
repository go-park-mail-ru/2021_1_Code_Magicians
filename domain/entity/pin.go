package entity

import "sync"

type PinsStorage struct {
	Storage *PinSet
}

func NewPinsSet() *PinSet {
	return &PinSet{
		Mutex:    sync.RWMutex{},
		UserPins: map[int][]*Pin{},
		AllPins:  []*Pin{},
	}
}

type PinSet struct {
	UserPins map[int][]*Pin
	AllPins  []*Pin
	UserId   int
	PinId    int
	Mutex    sync.RWMutex
}

type Pin struct {
	PinId       int    `json:"id"`
	BoardID     int    `json:"boardID"`
	Title       string `json:"title"`
	ImageLink   string `json:"pinImage"`
	Description string `json:"description"`
}
