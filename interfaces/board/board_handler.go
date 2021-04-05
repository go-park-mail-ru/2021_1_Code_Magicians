package board

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"pinterest/domain/entity"

	"github.com/gorilla/mux"
)

func (boardSet *entity.BoardSet) HandleAddBoard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currBoard := entity.Board{
		Id: boardSet.LastFreeBoardId,
	}

	err = json.Unmarshal(data, &currBoard)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := boardSet.LastFreeBoardId
	boardSet.LastFreeBoardId++

	boardInput := &entity.Board{
		Id:          id,
		Title:       currBoard.Title,
		Description: currBoard.Description,
	}
	boardSet.UserId = r.Context().Value("userID").(int)

	boardSet.Mutex.Lock()

	if boardSet.UsersInitialBoards[boardSet.UserId].isCreated == false {
		boardSet.UsersInitialBoards[boardSet.UserId].isCreated = true
		boardSet.UsersInitialBoards[boardSet.UserId].idBoard = boardSet.LastFreeBoardId
	}

	boardSet.UserBoards[boardSet.UserId] = append(boardSet.UserBoards[boardSet.UserId], boardInput)
	boardCopy := *boardInput
	boardSet.allBoards = append(boardSet.allBoards, &boardCopy)

	boardSet.Mutex.Unlock()

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

	boardSet.UserId = r.Context().Value("userID").(int)

	if boardId == boardSet.UsersInitialBoards[boardSet.UserId].idBoard {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardSet.Mutex.Lock()

	for _, b := range boardSet.allBoards {
		if b.Id == boardId && boardSet.CheckBoards(boardSet.UserId, boardId) { // Checking that board belongs this user
			*b = *boardSet.allBoards[len(boardSet.allBoards)-1]
			boardSet.allBoards = boardSet.allBoards[:len(boardSet.allBoards)-1]
			break
		}
	}

	for _, b := range boardSet.UserBoards[boardSet.UserId] {
		if b.Id == boardId {
			*b = *boardSet.UserBoards[boardSet.UserId][len(boardSet.allBoards)-1]
			boardSet.UserBoards[boardSet.UserId] = boardSet.UserBoards[boardSet.UserId][:len(boardSet.allBoards)-1]
			w.WriteHeader(http.StatusOK)
			boardSet.Mutex.Unlock()
			return
		}
	}

	boardSet.Mutex.Unlock()
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

	boardSet.Mutex.RLock()

	for _, b := range boardSet.allBoards {
		if b.Id == boardId {
			resultBoard = b
			break
		}
	}

	boardSet.Mutex.RUnlock()

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
	allBoards.Mutex.RLock()
	defer allBoards.Mutex.RUnlock()

	for _, boards := range allBoards.UserBoards[UseId] {
		if boards.Id == checkingId {
			return true
		}
	}
	return false
}
