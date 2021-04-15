package application

import (
	"encoding/json"
	"fmt"
	"log"
	"pinterest/domain/entity"
	"sync"

	"github.com/gorilla/websocket"
)

// TODO: renaming
type userNotificationsInfo struct {
	notifications map[int]entity.Notification
	client        *websocket.Conn
}

func newUserNotificationsInfo() *userNotificationsInfo {
	return &userNotificationsInfo{notifications: make(map[int]entity.Notification)}
}

type NotificationApp struct {
	notifications      map[int]userNotificationsInfo
	mu                 sync.Mutex
	lastNotificationID int
	userApp            UserAppInterface
}

func NewNotificationApp(userApp UserAppInterface) *NotificationApp {
	return &NotificationApp{
		notifications: make(map[int]userNotificationsInfo),
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
	ChangeClient(userID int, client *websocket.Conn) error
}

func (notificationApp *NotificationApp) AddNotification(notification *entity.Notification) (int, error) {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[notification.UserID]
	if !found {
		_, err := notificationApp.userApp.GetUser(notification.UserID)
		if err != nil {
			return 0, fmt.Errorf("User not found")
		}
		notificationsInfo = *newUserNotificationsInfo()
	}

	notification.NotificationID = notificationApp.lastNotificationID
	notificationApp.lastNotificationID++

	notificationsInfo.notifications[notification.NotificationID] = *notification
	notificationApp.notifications[notification.UserID] = notificationsInfo
	return notification.NotificationID, nil
}

func (notificationApp *NotificationApp) RemoveNotification(userID int, notificationID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}

	_, found = notificationsInfo.notifications[notificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	delete(notificationsInfo.notifications, notificationID)
	notificationApp.notifications[userID] = notificationsInfo
	return nil
}

func (notificationApp *NotificationApp) EditNotification(notification *entity.Notification) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[notification.UserID]
	if !found {
		return fmt.Errorf("User not found")
	}

	_, found = notificationsInfo.notifications[notification.NotificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	notificationApp.notifications[notification.UserID].notifications[notification.NotificationID] = *notification
	return nil
}

func (notificationApp *NotificationApp) GetNotification(userID int, notificationID int) (*entity.Notification, error) {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[userID]
	if !found {
		return nil, fmt.Errorf("User not found")
	}

	notification, found := notificationsInfo.notifications[notificationID]
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

	notificationsInfo, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}
	if notificationsInfo.client == nil {
		return fmt.Errorf("Notifications client is not set")
	}

	allNotifications := entity.MessageManyNotifications{Type: entity.AllNotificationsTypeKey, Notifications: make([]entity.Notification, 0)}

	for _, notification := range notificationsInfo.notifications {
		allNotifications.Notifications = append(allNotifications.Notifications, notification)
	}

	msg, err := json.Marshal(allNotifications)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not parse messages into JSON")
	}

	sendMessage(notificationsInfo.client, msg)

	return nil
}

func (notificationApp *NotificationApp) SendNotification(userID int, notificationID int) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[userID]
	if !found {
		return fmt.Errorf("User not found")
	}
	if notificationsInfo.client == nil {
		return fmt.Errorf("Notifications client is not set")
	}

	notification, found := notificationsInfo.notifications[notificationID]
	if !found {
		return fmt.Errorf("Notification not found")
	}

	notificationMsg := entity.MessageOneNotification{Type: entity.OneNotificationTypeKey, Notification: notification}

	msg, err := json.Marshal(notificationMsg)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not parse messages into JSON")
	}

	sendMessage(notificationsInfo.client, msg)

	return nil
}

func (notificationApp *NotificationApp) ChangeClient(userID int, client *websocket.Conn) error {
	notificationApp.mu.Lock()
	defer notificationApp.mu.Unlock()

	notificationsInfo, found := notificationApp.notifications[userID]
	if !found {
		_, err := notificationApp.userApp.GetUser(userID)
		if err != nil {
			return fmt.Errorf("User not found")
		}
		notificationsInfo = *newUserNotificationsInfo()
	}

	if notificationsInfo.client != nil {
		notificationsInfo.client.Close()
	}

	notificationsInfo.client = client
	notificationApp.notifications[userID] = notificationsInfo
	return nil
}
