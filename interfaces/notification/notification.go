package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"

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
		return
	}

	err = notificationInfo.notificationsApp.ChangeClient(initialMessage.UserID, ws)
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

	log.Println("Connected")
}
