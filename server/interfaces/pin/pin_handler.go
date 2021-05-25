package pin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"pinterest/domain/entity"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"pinterest/application"

	"github.com/gorilla/mux"
)

type PinInfo struct {
	pinApp   application.PinAppInterface
	boardApp application.BoardAppInterface
	s3App    application.S3AppInterface
	logger   *zap.Logger
}

func NewPinInfo(pinApp application.PinAppInterface,
	s3App application.S3AppInterface,
	boardApp application.BoardAppInterface,
	logger *zap.Logger) *PinInfo {
	return &PinInfo{
		pinApp:   pinApp,
		boardApp: boardApp,
		s3App:    s3App,
		logger:   logger,
	}
}

const maxPostPictureBodySize int = 8 * 1024 * 1024 // 8 mB

func (pinInfo *PinInfo) HandleAddPin(w http.ResponseWriter, r *http.Request) {
	bodySize := r.ContentLength
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	if bodySize < 0 { // No picture was passed
		pinInfo.logger.Info(entity.NoPicturePassed.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if bodySize > int64(maxPostPictureBodySize) { // Picture is too large
		pinInfo.logger.Info(entity.TooLargePicture.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(bodySize)
	jsonData := r.FormValue(string(entity.PinInfoLabelKey))
	currPin := entity.Pin{}
	err := json.Unmarshal([]byte(jsonData), &currPin)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	currPin.UserID = userID

	file, header, err := r.FormFile(string(entity.PinImageLabelKey))
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	extension := filepath.Ext(header.Filename)

	currPin.PinID, err = pinInfo.pinApp.CreatePin(&currPin, file, extension)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		switch err {
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pinIDOutput := entity.PinID{currPin.PinID}
	body, err := json.Marshal(pinIDOutput)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		pinInfo.pinApp.DeletePin(currPin.PinID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (pinInfo *PinInfo) HandleAddPinToBoard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	err = pinInfo.boardApp.CheckBoard(userID, boardID)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		switch err {
		case entity.CheckBoardOwnerError:
			w.WriteHeader(http.StatusForbidden)
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pinID, err := strconv.Atoi(vars[string(entity.PinIDLabelKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = pinInfo.pinApp.AddPin(boardID, pinID)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (pinInfo *PinInfo) HandleSavePin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pinID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	err = pinInfo.pinApp.SavePin(userID, pinID)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (pinInfo *PinInfo) HandleDelPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err = pinInfo.boardApp.CheckBoard(userID, boardID)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		switch err {
		case entity.CheckBoardOwnerError:
			w.WriteHeader(http.StatusForbidden)
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pinID, err := strconv.Atoi(vars[string(entity.PinIDLabelKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = pinInfo.pinApp.RemovePin(boardID, pinID)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (pinInfo *PinInfo) HandleGetPinByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pinID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultPin, err := pinInfo.pinApp.GetPin(pinID)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.PinNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	body, err := json.Marshal(resultPin)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (pinInfo *PinInfo) HandleGetPinsByBoardID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID, err := strconv.Atoi(vars[string(entity.IDKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	boardPins, err := pinInfo.pinApp.GetPins(boardID)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if len(boardPins) == 0 {
		pinInfo.logger.Info(entity.NoResultSearch.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range boardPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	pinsBody, err := json.Marshal(pins)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(pinsBody)
}

func (pinInfo *PinInfo) HandlePinsFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	numOfPins, err := strconv.Atoi(vars[string(entity.PinAmountLabelKey)])
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feedPins, err := pinInfo.pinApp.GetNumOfPins(numOfPins)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.NonPositiveNumOfPinsError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range feedPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	pinsBody, err := json.Marshal(pins)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(pinsBody)
}

func (pinInfo *PinInfo) HandleSearchPins(w http.ResponseWriter, r *http.Request) {
	keyString := mux.Vars(r)[string(entity.SearchKeyQuery)]

	keyString = strings.NewReplacer("+", " ").Replace(keyString)

	resultPins, err := pinInfo.pinApp.SearchPins(strings.ToLower(keyString))
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(resultPins) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range resultPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	responseBody, err := json.Marshal(pins)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (pinInfo *PinInfo) HandleCreateReport(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	report := new(entity.Report)

	err = json.Unmarshal(data, report)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID

	report.SenderID = userID
	report.ReportID, err = pinInfo.pinApp.CreateReport(report)
	if err != nil {
		pinInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.Int("for user", userID), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reportIDOutput := entity.ReportID{report.ReportID}
	body, err := json.Marshal(reportIDOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}
