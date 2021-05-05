package application

import (
	"fmt"
	"pinterest/domain/entity"
	"sync"

	"github.com/gorilla/websocket"
)

type websocketInfo struct {
	csrfToken string
	mu        sync.Mutex
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
	SendMessage(userID int, message []byte) error          // Send message to specified user (concurrency-safe)
	SendMessages(userID int, messages [][]byte) error      // Send messages to specified user (concurrency-safe)
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

func (websocketApp *WebsocketApp) getConnection(userID int) (*websocketInfo, error) {
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		_, err := websocketApp.userApp.GetUser(userID)
		if err != nil {
			return nil, entity.UserNotFoundError
		}

		return nil, entity.ClientNotSetError
	}

	return &connection, nil
}

func (websocketApp *WebsocketApp) GetClient(userID int) (*websocket.Conn, error) {
	connection, err := websocketApp.getConnection(userID)
	if err != nil {
		return nil, err
	}

	return connection.client, nil
}

func (websocketApp *WebsocketApp) ChangeToken(userID int, csrfToken string) error {
	connection, err := websocketApp.getConnection(userID)
	if err != nil {
		switch err {
		case entity.ClientNotSetError:
			connection = &websocketInfo{}
		default:
			return err
		}
	}

	connection.csrfToken = csrfToken
	websocketApp.mu.Lock()
	websocketApp.connections[userID] = *connection
	websocketApp.mu.Unlock()
	return nil
}

func (websocketApp *WebsocketApp) CheckToken(userID int, csrfToken string) error {
	connection, err := websocketApp.getConnection(userID)
	if err != nil {
		return err
	}

	if connection.csrfToken != csrfToken {
		return fmt.Errorf("Incorrect CSRF token")
	}

	return nil
}

func sendMessage(client *websocket.Conn, message []byte) error { // Is not safe for concurrent use
	w, err := client.NextWriter(websocket.TextMessage)
	if err != nil {
		return fmt.Errorf("Could not start writing")
	}

	w.Write(message)
	w.Close()
	return nil
}

func (websocketApp *WebsocketApp) SendMessage(userID int, message []byte) error {
	connection, err := websocketApp.getConnection(userID)
	if err != nil {
		return err
	}

	connection.mu.Lock()
	defer connection.mu.Unlock()

	if connection.client == nil {
		return entity.ClientNotSetError
	}

	err = sendMessage(connection.client, message)
	if err != nil {
		return err
	}

	return nil
}

func (websocketApp *WebsocketApp) SendMessages(userID int, messages [][]byte) error {
	connection, err := websocketApp.getConnection(userID)
	if err != nil {
		return err
	}

	connection.mu.Lock()
	defer connection.mu.Unlock()

	if connection.client == nil {
		return entity.ClientNotSetError
	}

	for _, message := range messages {
		err := sendMessage(connection.client, message)
		if err != nil {
			return err
		}
	}

	return nil
}
