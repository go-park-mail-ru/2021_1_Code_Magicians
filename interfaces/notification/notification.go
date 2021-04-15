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
	userID := 74
	err = notificationInfo.notificationsApp.ChangeClient(userID, ws)
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	err = notificationInfo.notificationsApp.SendAllNotifications(userID)
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	log.Println("Connected")
}
