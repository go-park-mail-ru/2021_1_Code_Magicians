package notification

import (
	"encoding/json"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type NotificationInfo struct {
	notificationsApp application.NotificationAppInterface
	websocketApp     application.WebsocketAppInterface
	csrfOn           bool
	logger           *zap.Logger
}

func NewNotificationInfo(notificationsApp application.NotificationAppInterface,
	csrfOn bool, logger *zap.Logger) *NotificationInfo {
	return &NotificationInfo{
		notificationsApp: notificationsApp,
		csrfOn:           csrfOn,
		logger:           logger,
	}
}

func (notificationInfo *NotificationInfo) HandleConnect(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024 * 1024,
		WriteBufferSize: 1024 * 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		notificationInfo.logger.Info(err.Error())
		return
	}

	_, initialMessageBytes, err := ws.ReadMessage() // TODO: add timeout
	if err != nil {
		notificationInfo.logger.Info(err.Error())
		return
	}
	var initialMessage entity.InitialMessage
	err = json.Unmarshal(initialMessageBytes, &initialMessage)
	if err != nil {
		notificationInfo.logger.Info(err.Error())
		return
	}

	if notificationInfo.csrfOn {
		err = notificationInfo.websocketApp.CheckToken(initialMessage.UserID, initialMessage.CSRFToken)
		if err != nil {
			notificationInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
			ws.Close()
			return
		}
	}

	err = notificationInfo.websocketApp.ChangeClient(initialMessage.UserID, ws)
	if err != nil {
		notificationInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
		ws.Close()
		return
	}

	err = notificationInfo.notificationsApp.SendAllNotifications(initialMessage.UserID)
	if err != nil {
		notificationInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
		ws.Close()
		return
	}
}

func (notificationInfo *NotificationInfo) HandleReadNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationIDStr := vars[string(entity.IDKey)]
	notificationID, _ := strconv.Atoi(notificationIDStr)
	userID := r.Context().Value(entity.CookieInfoKey).(*entity.CookieInfo).UserID
	err := notificationInfo.notificationsApp.ReadNotification(userID, notificationID)

	if err != nil {
		notificationInfo.logger.Info(err.Error(), zap.Int("for user", userID))
		switch err.Error() {
		case "Notification not found":
			w.WriteHeader(http.StatusNotFound)
		case "Notification was already read":
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
