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
	BoardApp application.BoardAppInterface
}

func (boardInfo *BoardInfo) HandleAddBoard(w http.ResponseWriter, r *http.Request) {
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

	boardInput := &entity.Board{
		Title:       currBoard.Title,
		Description: currBoard.Description,
	}
	boardInput.UserID, err = boardInfo.BoardApp.AddBoard(boardInput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := `{"board_id": ` + strconv.Itoa(boardInput.BoardID) + `}`

	w.WriteHeader(http.StatusCreated) // returning success code
	w.Write([]byte(body))
}

func (boardInfo *BoardInfo) HandleDelBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("cookieInfo").(*entity.CookieInfo).UserID
	err = boardInfo.BoardApp.DeleteBoard(boardId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (boardInfo *BoardInfo) HandleGetBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoard, err := boardInfo.BoardApp.GetBoard(boardId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (boardInfo *BoardInfo) HandleGetBoardsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoards, err := boardInfo.BoardApp.GetBoards(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(resultBoards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
