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
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		return entity.ClientNotSetError
	}

	connection.mu.Lock()
	defer connection.mu.Unlock()

	err := sendMessage(connection.client, message)
	if err != nil {
		return err
	}

	return nil
}

func (websocketApp *WebsocketApp) SendMessages(userID int, messages [][]byte) error {
	websocketApp.mu.Lock()
	defer websocketApp.mu.Unlock()

	connection, found := websocketApp.connections[userID]
	if !found {
		return entity.ClientNotSetError
	}

	connection.mu.Lock()
	defer connection.mu.Unlock()

	for _, message := range messages {
		err := sendMessage(connection.client, message)
		if err != nil {
			return err
		}
	}

	return nil
}
