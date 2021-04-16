package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type NotificationInfo struct {
	notificationsApp application.NotificationsAppInterface
}

func NewNotificationInfo(notificationsApp application.NotificationsAppInterface) *NotificationInfo {
	return &NotificationInfo{notificationsApp: notificationsApp}
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
		log.Println(err)
		return
	}

	_, initialMessageBytes, err := ws.ReadMessage() // TODO: add timeout
	if err != nil {
		log.Println(err)
		return
	}
	var initialMessage entity.InitialMessage
	err = json.Unmarshal(initialMessageBytes, &initialMessage)
	if err != nil {
		log.Println(err)
		return
	}

	err = notificationInfo.notificationsApp.ChangeClient(initialMessage.UserID, ws, "")
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	err = notificationInfo.notificationsApp.SendAllNotifications(initialMessage.UserID)
	if err != nil {
		log.Println(err)
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
