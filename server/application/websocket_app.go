package application

import (
	"fmt"
	"pinterest/domain/entity"
	"sync"

	"github.com/gorilla/websocket"
)

type websocketInfo struct {
	csrfToken string
	client    *websocket.Conn
}

type WebsocketApp struct {
	connections map[int]websocketInfo
	mu          sync.Mutex
	userApp     UserAppInterface
}

func NewWebsocketApp(userApp UserAppInterface) *WebsocketApp {
	return &WebsocketApp{
		connections: make(map[int]websocketInfo),
		userApp:     userApp,
	}
}

type WebsocketAppInterface interface {
	ChangeClient(userID int, client *websocket.Conn) error // Switches client  that was assigned to user
	GetClient(userID int) (*websocket.Conn, error)         // Get user's client
	ChangeToken(userID int, csrfToken string) error        // Change user's CRSF token
	CheckToken(userID int, csrfToken string) error         // Check if passed token is correct (nil on success)
}

func (websocketApp *WebsocketApp) ChangeClient(userID int, client *websocket.Conn) error {
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		_, err := websocketApp.userApp.GetUser(userID)
		if err != nil {
			return entity.UserNotFoundError
		}

		connection = websocketInfo{}
	}

	if connection.client != nil {
		connection.client.Close()
	}

	connection.client = client
	websocketApp.connections[userID] = connection
	return nil
}

func (websocketApp *WebsocketApp) GetClient(userID int) (*websocket.Conn, error) {
	connection, found := websocketApp.connections[userID]
	if !found {
		return nil, entity.ClientNotSetError
	}

	return connection.client, nil
}

func (websocketApp *WebsocketApp) ChangeToken(userID int, csrfToken string) error {
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		_, err := websocketApp.userApp.GetUser(userID)
		if err != nil {
			return entity.UserNotFoundError
		}

		connection = websocketInfo{}
	}

	connection.csrfToken = csrfToken
	websocketApp.connections[userID] = connection
	return nil
}

func (websocketApp *WebsocketApp) CheckToken(userID int, csrfToken string) error {
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		_, err := websocketApp.userApp.GetUser(userID)
		if err != nil {
			return entity.UserNotFoundError
		}

		connection = websocketInfo{}
	}

	if connection.csrfToken != csrfToken {
		return fmt.Errorf("Incorrect CSRF token")
	}

	return nil
}
