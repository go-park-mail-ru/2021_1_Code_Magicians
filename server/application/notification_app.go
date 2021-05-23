package application

import (
	"encoding/json"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type NotificationApp struct {
	notificationRepo repository.NotificationRepository
	userApp          UserAppInterface
	websocketApp     WebsocketAppInterface
}

func NewNotificationApp(notificationRepo repository.NotificationRepository, userApp UserAppInterface, websocketApp WebsocketAppInterface) *NotificationApp {
	return &NotificationApp{
		notificationRepo: notificationRepo,
		userApp:          userApp,
		websocketApp:     websocketApp,
	}
}

type NotificationAppInterface interface {
	AddNotification(notification *entity.Notification) (int, error)               // Add notification to list of user's notifications
	RemoveNotification(userID int, notificationID int) error                      // Remove notification from list of user's notifications
	EditNotification(notification *entity.Notification) error                     // Change fields of notification with same user and notification ID
	GetNotification(userID int, notificationID int) (*entity.Notification, error) // Get notification from db using user's and notification's IDs
	SendAllNotifications(userID int) error                                        // Send all of the notifications that this user has
	SendNotification(userID int, notificationID int) error                        // Send specified  notification to specified user
	ReadNotification(userID int, notificationID int) error                        // Changes notification's status to "Read"
}

func (notificationApp *NotificationApp) AddNotification(notification *entity.Notification) (int, error) {
	return notificationApp.notificationRepo.AddNotification(notification)
}

func (notificationApp *NotificationApp) RemoveNotification(userID int, notificationID int) error {
	notification, err := notificationApp.notificationRepo.GetNotification(notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return entity.ForeignNotificationError
	}

	return notificationApp.notificationRepo.RemoveNotification(notificationID)
}

func (notificationApp *NotificationApp) EditNotification(notification *entity.Notification) error {
	oldNotification, err := notificationApp.notificationRepo.GetNotification(notification.NotificationID)
	if err != nil {
		return err
	}

	if oldNotification.UserID != notification.UserID {
		return entity.ForeignNotificationError
	}

	return notificationApp.notificationRepo.EditNotification(notification)
}

func (notificationApp *NotificationApp) GetNotification(userID int, notificationID int) (*entity.Notification, error) {
	notification, err := notificationApp.notificationRepo.GetNotification(notificationID)
	if err != nil {
		return nil, err
	}

	if notification.UserID != userID {
		return nil, entity.ForeignNotificationError
	}
	return notification, nil
}

func (notificationApp *NotificationApp) SendAllNotifications(userID int) error {
	notifications, err := notificationApp.notificationRepo.GetAllNotifications(userID)
	if err != nil {
		switch err {
		case entity.NotificationsNotFoundError:
			notifications = make([]*entity.Notification, 0)
		default:
			return err
		}
	}

	notificationsOutput := entity.AllNotificationsOutput{
		Type:          entity.AllNotificationsTypeKey,
		Notifications: make([]entity.Notification, 0, len(notifications)),
	}

	for _, notification := range notifications {
		notificationsOutput.Notifications = append(notificationsOutput.Notifications, *notification)
	}

	message, err := json.Marshal(notificationsOutput)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = notificationApp.websocketApp.SendMessage(userID, message)

	return err
}

func (notificationApp *NotificationApp) SendNotification(userID int, notificationID int) error {
	notification, err := notificationApp.GetNotification(userID, notificationID)
	if err != nil {
		return err
	}

	notificationOutput := entity.OneNotificationOutput{Type: entity.OneNotificationTypeKey, Notification: *notification}

	message, err := json.Marshal(notificationOutput)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = notificationApp.websocketApp.SendMessage(userID, message)

	return err
}

func (notificationApp *NotificationApp) ReadNotification(userID int, notificationID int) error {
	notification, err := notificationApp.GetNotification(userID, notificationID)
	if err != nil {
		return err
	}

	if notification.IsRead {
		return entity.NotificationAlreadyReadError
	}

	notification.IsRead = true
	return notificationApp.notificationRepo.EditNotification(notification)
}
