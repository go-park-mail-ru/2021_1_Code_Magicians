package board

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"pinterest/domain/entity"

	"pinterest/application"

	"github.com/gorilla/mux"
)

type BoardInfo struct {
	boardApp application.BoardAppInterface
	logger   *zap.Logger
}

func NewBoardInfo(boardApp application.BoardAppInterface, logger *zap.Logger) *BoardInfo {
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
	boardInput.BoardID, err = boardInfo.boardApp.CreateBoard(boardInput)
	if err != nil {
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardID := entity.BoardID{BoardID: boardInput.BoardID}
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
		switch err {
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		case entity.CheckBoardOwnerError:
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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
		switch err {
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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
	userID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultBoards, err := boardInfo.boardApp.GetBoards(userID)
	if err != nil && err != entity.BoardsNotFoundError { // It's fine if no boards were found
		boardInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if resultBoards == nil {
		resultBoards = make([]entity.Board, 0) // So that [] appears in json and not nil
	}

	boards := entity.BoardsOutput{Boards: resultBoards}

	boardsBody, err := json.Marshal(boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(boardsBody)
}
