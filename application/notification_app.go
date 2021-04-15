package application

import (
	"pinterest/domain/entity"
	"sync"
)

type NotificationApp struct {
	notifications map[int]map[int]entity.Notification
	mu            sync.Mutex
}

func NewNotificationApp() *NotificationApp {
	return &NotificationApp{
		notifications: make(map[int]map[int]entity.Notification),
	}
}

type NotificationsAppInterface interface {
	AddNotification(*entity.Notification) (int, error)      // Add notification to list of user's notifications
	RemoveNotification(*entity.Notification) error          // Remove notification from list of user's notifications
	EditNotification(*entity.Notification) error            // Change fields of notification in database (except for ids)
	GetNotification(int, int) (*entity.Notification, error) // Get notification from db using user's and notification's IDs
}

func (notificationsApp *NotificationApp) AddNotification(notification *entity.Notification) (int, error) {
	return 0, nil
}
func (notificationsApp *NotificationApp) RemoveNotification(notification *entity.Notification) error {
	return nil
}
func (notificationsApp *NotificationApp) EditNotification(notification *entity.Notification) error {
	return nil
}
func (notificationsApp *NotificationApp) GetNotification(userID int, notificationID int) (*entity.Notification, error) {
	return nil, nil
}
