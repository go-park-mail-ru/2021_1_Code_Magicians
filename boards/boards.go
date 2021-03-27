package boards

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

func NewBoardSet() *BoardSet {
	return &BoardSet{
		mutex:      sync.RWMutex{},
		userBoards: map[int][]*Board{},
		allBoards:  []*Board{},
	}
}

type BoardSet struct {
	userBoards map[int][]*Board
	allBoards  []*Board
	//pinsInBoard map[int][]*Pin
	userId  int
	BoardId int
	mutex   sync.RWMutex
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
		Id: boardSet.BoardId,
	}

	err = json.Unmarshal(data, &currBoard)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := boardSet.BoardId
	boardSet.BoardId++

	boardInput := &Board{
		Id:          id,
		Title:       currBoard.Title,
		Description: currBoard.Description,
	}

	boardSet.mutex.Lock()

	boardSet.userId = r.Context().Value("userID").(int)

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

	if boardId > boardSet.allBoards[len(boardSet.allBoards) - 1].Id {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	boardSet.userId = r.Context().Value("userID").(int)



	boardSet.mutex.Lock()

	for _, b := range boardSet.allBoards {
		if b.Id == boardId {
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

func (boardSet *BoardSet) HandleGetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if boardId > boardSet.allBoards[len(boardSet.allBoards) - 1].Id {
		w.WriteHeader(http.StatusNotFound)
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
