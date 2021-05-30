package pin

import (
	"encoding/json"
	"fmt"
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
	pinApp          application.PinAppInterface
	followApp       application.FollowAppInterface
	notificationApp application.NotificationAppInterface
	userApp         application.UserAppInterface
	boardApp        application.BoardAppInterface
	s3App           application.S3AppInterface
	logger          *zap.Logger
	template        string // Used for creating an e-mail for notifications
	emailUsername   string
	emailPassword   string
}

func NewPinInfo(pinApp application.PinAppInterface, followApp application.FollowAppInterface,
	notificationApp application.NotificationAppInterface, userApp application.UserAppInterface,
	boardApp application.BoardAppInterface, s3App application.S3AppInterface,
	logger *zap.Logger, template string, emailUsername string, emailPassword string) *PinInfo {
	return &PinInfo{
		pinApp:          pinApp,
		followApp:       followApp,
		notificationApp: notificationApp,
		userApp:         userApp,
		boardApp:        boardApp,
		s3App:           s3App,
		logger:          logger,
		template:        template,
		emailUsername:   emailUsername,
		emailPassword:   emailPassword,
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

	user, err := pinInfo.userApp.GetUser(userID)
	if err != nil {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		pinInfo.pinApp.DeletePin(currPin.PinID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go pinInfo.sendNotificationsAndEmails(user, currPin)

	pinIDOutput := entity.PinID{PinID: currPin.PinID}
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

func (pinInfo *PinInfo) sendNotificationsAndEmails(sender *entity.User, pin entity.Pin) {
	var usersWithNotifications []entity.UserNotificationInfo
	var err error

	followers, err := pinInfo.followApp.GetAllFollowers(sender.UserID)
	if err != nil {
		return
	}

	for _, user := range followers {
		notification := entity.Notification{
			UserID:   user.UserID,
			Title:    "New Pin from people you've subscribed to!",
			Category: "subscribed pins",
			Text: fmt.Sprintf(`%s! You have a new pin from user %s: "%s"`,
				user.Username, sender.Username, pin.Title),
			IsRead: false,
		}
		notification.NotificationID, err = pinInfo.notificationApp.AddNotification(&notification)

		switch err {
		case nil:
			usersWithNotifications = append(usersWithNotifications, entity.UserNotificationInfo{
				UserID:         user.UserID,
				NotificationID: notification.NotificationID,
			})
		default:
			pinInfo.logger.Info(err.Error(), zap.String("function", "PinInfo.sendNotificationsAndEmails"),
				zap.Int("for user", user.UserID))
		}
	}

	go pinInfo.notificationApp.SendNotificationsToUsers(usersWithNotifications)

	go pinInfo.sendEmails(usersWithNotifications, pin.PinID)
}

func (pinInfo *PinInfo) sendEmails(usersAndNotifications []entity.UserNotificationInfo, pinID int) {
	for _, pair := range usersAndNotifications {
		user, err := pinInfo.userApp.GetUser(pair.UserID)
		if err != nil {
			pinInfo.logger.Info(err.Error(), zap.String("function", "PinInfo.sendEmails"),
				zap.Int("for user", user.UserID))
			continue
		}
		notification, err := pinInfo.notificationApp.GetNotification(pair.UserID, pair.NotificationID)
		if err != nil {
			pinInfo.logger.Info(err.Error(), zap.String("function", "PinInfo.sendEmails"),
				zap.Int("for user", user.UserID))
			continue
		}

		templateStruct := struct {
			NotificationTitle    string
			NotificationText     string
			NotificationID       int
			NotificationCategory string
			Username             string
			PinID                int
		}{
			NotificationTitle:    notification.Title,
			NotificationText:     notification.Text,
			NotificationID:       notification.NotificationID,
			NotificationCategory: notification.Category,
			Username:             user.Username,
			PinID:                pinID,
		}
		err = pinInfo.notificationApp.SendNotificationEmail(user.UserID, notification.NotificationID,
			pinInfo.template, templateStruct,
			pinInfo.emailUsername, pinInfo.emailPassword)
		if err != nil {
			pinInfo.logger.Info(err.Error(), zap.String("function", "PinInfo.sendEmails"),
				zap.Int("for user", user.UserID))
		}
	}
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
	if err != nil && err != entity.PinsNotFoundError {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.BoardNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range boardPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	if pins.Pins == nil {
		pins.Pins = make([]entity.PinOutput, 0) // So that [] appears in json and not nil
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
	feedInfo := new(entity.FeedInfo)
	queryParams := r.URL.Query()

	offsetList, exists := queryParams["offset"]
	if !exists {
		pinInfo.logger.Info("offset was not passed", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var err error
	feedInfo.Offset, err = strconv.Atoi(offsetList[0])
	if err != nil {
		pinInfo.logger.Info("offset is not a number", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if feedInfo.Offset < 0 {
		pinInfo.logger.Info("offset cannot be negative", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	amountList, exists := queryParams["amount"]
	if !exists {
		pinInfo.logger.Info("amount was not passed", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	feedInfo.Amount, err = strconv.Atoi(amountList[0])
	if err != nil {
		pinInfo.logger.Info("amount is not a number", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if feedInfo.Amount < 0 {
		pinInfo.logger.Info("amount cannot be negative", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feedPins, err := pinInfo.pinApp.GetPinsWithOffset(feedInfo.Offset, feedInfo.Amount)
	if err != nil && err != entity.PinsNotFoundError {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		switch err {
		case entity.PinScanError:
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

	if pins.Pins == nil {
		pins.Pins = make([]entity.PinOutput, 0) // So that [] appears in json and not nil
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
	queryParams := r.URL.Query()

	keywordsList, exists := queryParams["searchKey"]
	if !exists {
		pinInfo.logger.Info("searchKey was not passed", zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	keyWords := keywordsList[0]
	interval := ""
	intervalsList, exists := queryParams["interval"]
	switch exists {
	case true:
		switch intervalsList[0] {
		case "day", "week", "hour", "allTime":
			interval = intervalsList[0]
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case false:
		interval = "allTime"
	}

	keyWords = strings.NewReplacer("+", " ").Replace(keyWords)

	resultPins, err := pinInfo.pinApp.SearchPins(strings.ToLower(keyWords), interval)
	if err != nil && err != entity.PinsNotFoundError {
		pinInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pins := new(entity.PinsListOutput)

	for _, pin := range resultPins {
		var pinOutput entity.PinOutput
		pinOutput.FillFromPin(&pin)
		pins.Pins = append(pins.Pins, pinOutput)
	}

	if pins.Pins == nil {
		pins.Pins = make([]entity.PinOutput, 0) // So that [] appears in json and not nil
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
		switch err {
		case entity.DuplicateReportError:
			w.WriteHeader(http.StatusConflict)
		case entity.PinNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			pinInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.Int("for user", userID), zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	reportIDOutput := entity.ReportID{ReportID: report.ReportID}
	body, err := json.Marshal(reportIDOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}
