package application

import (
	"encoding/json"
	"fmt"
	"log"
	"pinterest/domain/entity"
	"sync"

	"github.com/gorilla/websocket"
)

type connectionInfo struct {
	csrfToken string
	client    *websocket.Conn
}

type NotificationApp struct {
	notifications      map[int]map[int]entity.Notification
	lastNotificationID int
	connections        map[int]connectionInfo
	mu                 sync.Mutex
	userApp            UserAppInterface
}

func NewNotificationApp(userApp UserAppInterface) *NotificationApp {
	return &NotificationApp{
		notifications: make(map[int]map[int]entity.Notification),
		connections:   make(map[int]connectionInfo),
		userApp:       userApp,
	}
}

type NotificationsAppInterface interface {
	AddNotification(notification *entity.Notification) (int, error)               // Add notification to list of user's notifications
	RemoveNotification(userID int, notificationID int) error                      // Remove notification from list of user's notifications
	EditNotification(notification *entity.Notification) error                     // Change fields of notification with same user and notification ID
	GetNotification(userID int, notificationID int) (*entity.Notification, error) // Get notification from db using user's and notification's IDs
	SendAllNotifications(userID int) error                                        // Send all of the notifications that this user has
	SendNotification(userID int, notificationID int) error                        // Send specified  notification to specified user
	ChangeClient(userID int, client *websocket.Conn, csrfToken string) error      // Switches client  that was assigned to user
	ReadNotification(userID int, notificationID int) error                        // Changes notification's status to "Read"
}

func (notificationApp *NotificationApp) AddNotification(notification *entity.Notification) (int, error) {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[notification.UserID]
	if !found {
		_, err := notificationApp.userApp.GetUser(notification.UserID)
		if err != nil {
			return 0, fmt.Errorf("User not found")
		}
		notificationsMap = make(map[int]entity.Notification)
	}

	notification.NotificationID = notificationApp.lastNotificationID
	notificationApp.lastNotificationID++

	notificationsMap[notification.NotificationID] = *notification
	notificationApp.notifications[notification.UserID] = notificationsMap
	return notification.NotificationID, nil
}

func (notificationApp *NotificationApp) RemoveNotification(userID int, notificationID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}
	if notificationsMap == nil {
		return fmt.Errorf("User has no notifications")
	}

	_, found = notificationsMap[notificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	delete(notificationsMap, notificationID)
	notificationApp.notifications[userID] = notificationsMap
	return nil
}

func (notificationApp *NotificationApp) EditNotification(notification *entity.Notification) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[notification.UserID]
	if !found {
		return fmt.Errorf("User not found")
	}
	if notificationsMap == nil {
		return fmt.Errorf("User has no notifications")
	}

	_, found = notificationsMap[notification.NotificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	notificationApp.notifications[notification.UserID][notification.NotificationID] = *notification
	return nil
}

func (notificationApp *NotificationApp) GetNotification(userID int, notificationID int) (*entity.Notification, error) {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[userID]
	if !found {
		return nil, fmt.Errorf("User not found")
	}
	if notificationsMap == nil {
		return nil, fmt.Errorf("User has no notifications")
	}

	notification, found := notificationsMap[notificationID]
	if !found {
		return nil, fmt.Errorf("Notification not found")
	}
	return &notification, nil
}

func sendMessage(client *websocket.Conn, msg []byte) error {
	w, err := client.NextWriter(websocket.TextMessage)
	if err != nil {
		return fmt.Errorf("Could not start writing")
	}

	w.Write(msg)
	w.Close()
	return nil
}

func (notificationApp *NotificationApp) SendAllNotifications(userID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}

	connection, found := notificationApp.connections[userID]
	if !found {
		return fmt.Errorf("Notifications client is not set")
	}
	// TODO: check csrf

	allNotifications := entity.MessageManyNotifications{Type: entity.AllNotificationsTypeKey, Notifications: make([]entity.Notification, 0)}

	for _, notification := range notificationsMap {
		allNotifications.Notifications = append(allNotifications.Notifications, notification)
	}

	msg, err := json.Marshal(allNotifications)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not parse messages into JSON")
	}

	sendMessage(connection.client, msg)

	return nil
}

func (notificationApp *NotificationApp) SendNotification(userID int, notificationID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}

	notification, found := notificationsMap[notificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	connection, found := notificationApp.connections[userID]
	if !found {
		return fmt.Errorf("Notifications client is not set")
	}
	// TODO: check csrf

	notificationMsg := entity.MessageOneNotification{Type: entity.OneNotificationTypeKey, Notification: notification}

	msg, err := json.Marshal(notificationMsg)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not parse messages into JSON")
	}

	sendMessage(connection.client, msg)

	return nil
}

func (notificationApp *NotificationApp) ChangeClient(userID int, client *websocket.Conn, csrfToken string) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	connection, found := notificationApp.connections[userID]
	if !found {
		_, err := notificationApp.userApp.GetUser(userID)
		if err != nil {
			return fmt.Errorf("User not found")
		}

		connection = connectionInfo{}
	}

	if connection.client != nil {
		connection.client.Close()
	}

	connection.csrfToken = csrfToken
	connection.client = client
	notificationApp.connections[userID] = connection
	return nil
}

func (notificationApp *NotificationApp) ReadNotification(userID int, notificationID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsMap, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}

	notification, found := notificationsMap[notificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	if notification.IsRead {
		return fmt.Errorf("Notification was already read")
	}

	notification.IsRead = true
	notificationsMap[notificationID] = notification
	notificationApp.notifications[userID] = notificationsMap
	return nil
}
