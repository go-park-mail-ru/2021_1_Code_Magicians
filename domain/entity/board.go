package entity

import "sync"

type InitialBoard struct {
	isCreated bool
	idBoard   int
}

func NewBoardSet() *BoardSet {
	return &BoardSet{
		mutex:              sync.RWMutex{},
		userBoards:         map[int][]*Board{},
		allBoards:          []*Board{},
		usersInitialBoards: map[int]*InitialBoard{},
	}
}

type BoardSet struct {
	userBoards         map[int][]*Board
	allBoards          []*Board
	usersInitialBoards map[int]*InitialBoard // Users default boards for all their pins
	userId             int
	LastFreeBoardId    int
	mutex              sync.RWMutex
}

type BoardsStorage struct {
	Storage *BoardSet
}

type Board struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
