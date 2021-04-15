package notification

import (
	"log"
	"net/http"
	"pinterest/application"

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

	// TODO: parse cxrf so that we know user's ID
	userID := 0
	notificationInfo.notificationsApp.ChangeClient(userID, ws)
	notificationInfo.notificationsApp.SendAllNotifications(userID)
}
