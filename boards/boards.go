package boards

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

type InitialBoard struct {
	isCreated bool
	idBoard int
}

func NewBoardSet() *BoardSet {

	return &BoardSet{
		mutex:      sync.RWMutex{},
		userBoards: map[int][]*Board{},
		allBoards:  []*Board{},
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

func (boardSet *BoardSet) HandleAddBoard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()


	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currBoard := Board{
		Id: boardSet.LastFreeBoardId,
	}

	err = json.Unmarshal(data, &currBoard)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	id := boardSet.LastFreeBoardId
	boardSet.LastFreeBoardId++

	boardInput := &Board{
		Id:          id,
		Title:       currBoard.Title,
		Description: currBoard.Description,
	}
	boardSet.userId = r.Context().Value("userID").(int)

	boardSet.mutex.Lock()

	if boardSet.usersInitialBoards[boardSet.userId].isCreated == false {
		boardSet.usersInitialBoards[boardSet.userId].isCreated = true
		boardSet.usersInitialBoards[boardSet.userId].idBoard = boardSet.LastFreeBoardId
	}

	boardSet.userBoards[boardSet.userId] = append(boardSet.userBoards[boardSet.userId], boardInput)
	boardCopy := *boardInput
	boardSet.allBoards = append(boardSet.allBoards, &boardCopy)

	boardSet.mutex.Unlock()

	body := `{"board_id": ` + strconv.Itoa(currBoard.Id) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (boardSet *BoardSet) HandleDelBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardSet.userId = r.Context().Value("userID").(int)

	if boardId == boardSet.usersInitialBoards[boardSet.userId].idBoard {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardSet.mutex.Lock()

	for _, b := range boardSet.allBoards {
		if b.Id == boardId && boardSet.CheckBoards(boardSet.userId, boardId) { // Checking that board belongs this user
			*b = *boardSet.allBoards[len(boardSet.allBoards)-1]
			boardSet.allBoards = boardSet.allBoards[:len(boardSet.allBoards)-1]
			break
		}
	}

	for _, b := range boardSet.userBoards[boardSet.userId] {
		if b.Id == boardId {
			*b = *boardSet.userBoards[boardSet.userId][len(boardSet.allBoards)-1]
			boardSet.userBoards[boardSet.userId] = boardSet.userBoards[boardSet.userId][:len(boardSet.allBoards)-1]
			w.WriteHeader(http.StatusOK)
			boardSet.mutex.Unlock()
			return
		}
	}

	boardSet.mutex.Unlock()
	w.WriteHeader(http.StatusNotFound)
}

func (boardSet *BoardSet) HandleGetBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var resultBoard *Board

	boardSet.mutex.RLock()

	for _, b := range boardSet.allBoards {
		if b.Id == boardId {
			resultBoard = b
			break
		}
	}

	boardSet.mutex.RUnlock()

	if resultBoard == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(resultBoard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (allBoards *BoardSet) CheckBoards(useId int, checkingId int) bool {
	allBoards.mutex.RLock()
	defer allBoards.mutex.RUnlock()

	for _, boards := range allBoards.userBoards[useId] {
		if boards.Id == checkingId  {
			return true
		}
	}
	return false
}