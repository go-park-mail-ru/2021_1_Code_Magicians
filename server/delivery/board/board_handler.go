package board

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"pinterest/domain/entity"

	"pinterest/usecase"

	"github.com/gorilla/mux"
)

type BoardInfo struct {
	boardApp usecase.BoardAppInterface
	logger   *zap.Logger
}

func NewBoardInfo(boardApp usecase.BoardAppInterface, logger *zap.Logger) *BoardInfo {
	return &BoardInfo{
		boardApp: boardApp,
		logger:   logger,
	}
}

func (boardInfo *BoardInfo) HandleCreateBoard(w http.ResponseWriter, r *http.Request) {
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
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardID := entity.BoardID{boardInput.BoardID}
	body, err := json.Marshal(boardID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "usecase/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (boardInfo *BoardInfo) HandleDelBoardByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = boardInfo.boardApp.DeleteBoard(userID, boardID)
	if err != nil {
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
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
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(resultBoard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "usecase/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (boardInfo *BoardInfo) HandleGetBoardsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoards, err := boardInfo.boardApp.GetBoards(userID)
	if err != nil {
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(resultBoards) == 0 {
		boardInfo.logger.Info(
			entity.NoResultSearch.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	boards := entity.BoardsOutput{resultBoards}

	boardsBody, err := json.Marshal(boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "usecase/json")
	w.WriteHeader(http.StatusOK)
	w.Write(boardsBody)
}
