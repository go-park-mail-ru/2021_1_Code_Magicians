package repository

import "pinterest/domain/entity"

type NotificationRepositoryInterface interface {
	AddNotification(notification *entity.Notification) (int, error)   // Add notification to database
	RemoveNotification(notificationID int) error                      // Remove notification from database
	EditNotification(notification *entity.Notification) error         // Change fields of notification with same notification ID
	GetNotification(notificationID int) (*entity.Notification, error) // Get notification from db using notification ID
	GetAllNotifications(userID int) ([]*entity.Notification, error)   // Get all notifications for specified user
}
