package board

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/application"
	"strconv"

	"pinterest/domain/entity"

	"github.com/gorilla/mux"
)

type BoardInfo struct {
	boardApp application.BoardAppInterface
}

func NewBoardInfo(boardApp application.BoardAppInterface) *BoardInfo {
	return &BoardInfo{
		boardApp: boardApp,
	}
}

func (boardInfo *BoardInfo) HandleCreateBoard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currBoard := entity.Board{}

	err = json.Unmarshal(data, &currBoard)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	boardInput := &entity.Board{
		UserID:      userID,
		Title:       currBoard.Title,
		Description: currBoard.Description,
	}
	boardInput.BoardID, err = boardInfo.boardApp.AddBoard(boardInput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	boardID := entity.BoardID{boardInput.BoardID}
	body, err := json.Marshal(boardID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (boardInfo *BoardInfo) HandleDelBoardByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = boardInfo.boardApp.DeleteBoard(boardId, userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (boardInfo *BoardInfo) HandleGetBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars[string(entity.IDKey)])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoard, err := boardInfo.boardApp.GetBoard(boardId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if resultBoard == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(resultBoard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (boardInfo *BoardInfo) HandleGetBoardsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars[string(entity.IDKey)])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoards, err := boardInfo.boardApp.GetBoards(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boards := entity.BoardsOutput{resultBoards}

	boardsBody, err := json.Marshal(boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(boardsBody)
}
