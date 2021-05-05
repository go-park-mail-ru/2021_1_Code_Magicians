package websocket

import (
	"encoding/json"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebsocketInfo struct {
	notificationApp application.NotificationAppInterface
	chatApp         application.ChatAppInterface
	websocketApp    application.WebsocketAppInterface
	csrfOn          bool
	logger          *zap.Logger
}

func NewWebsocketInfo(notificationApp application.NotificationAppInterface, chatApp application.ChatAppInterface,
	websocketApp application.WebsocketAppInterface,
	csrfOn bool, logger *zap.Logger) *WebsocketInfo {
	return &WebsocketInfo{
		notificationApp: notificationApp,
		chatApp:         chatApp,
		websocketApp:    websocketApp,
		csrfOn:          csrfOn,
		logger:          logger,
	}
}

func (websocketInfo *WebsocketInfo) HandleConnect(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024 * 1024,
		WriteBufferSize: 1024 * 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		websocketInfo.logger.Info(err.Error())
		return
	}

	_, initialMessageBytes, err := ws.ReadMessage() // TODO: add timeout
	if err != nil {
		websocketInfo.logger.Info(err.Error())
		ws.Close()
		return
	}
	var initialMessage entity.InitialMessage
	err = json.Unmarshal(initialMessageBytes, &initialMessage)
	if err != nil {
		websocketInfo.logger.Info(err.Error())
		ws.Close()
		return
	}

	if websocketInfo.csrfOn {
		err = websocketInfo.websocketApp.CheckToken(initialMessage.UserID, initialMessage.CSRFToken)
		if err != nil {
			websocketInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
			ws.Close()
			return
		}
	}

	err = websocketInfo.websocketApp.ChangeClient(initialMessage.UserID, ws)
	if err != nil {
		websocketInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
		ws.Close()
		return
	}

	err = websocketInfo.notificationApp.SendAllNotifications(initialMessage.UserID)
	if err != nil {
		websocketInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
		ws.Close()
		return
	}

	err = websocketInfo.chatApp.SendAllChats(initialMessage.UserID)
	if err != nil {
		websocketInfo.logger.Info(err.Error(), zap.Int("from user", initialMessage.UserID))
		ws.Close()
		return
	}
}
