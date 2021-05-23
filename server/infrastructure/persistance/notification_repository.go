package persistance

import (
	"pinterest/domain/entity"

	"github.com/tarantool/go-tarantool"
)

type NotificationRepo struct {
	tarantoolDB *tarantool.Connection
}

func NewNotificationRepository(tarantoolDB *tarantool.Connection) *NotificationRepo {
	return &NotificationRepo{tarantoolDB}
}

func (notificationRepo *NotificationRepo) AddNotification(notification *entity.Notification) (int, error) {
	notificationAsInterface := notificationToInterfaces(notification)
	resp, err := notificationRepo.tarantoolDB.Insert("notifications", notificationAsInterface)
	return 0, nil
}
func (notificationRepo *NotificationRepo) RemoveNotification(userID int, notificationID int) error {
	return nil
}
func (notificationRepo *NotificationRepo) EditNotification(notification *entity.Notification) error {
	return nil
}
func (notificationRepo *NotificationRepo) GetNotification(notificationID int) (*entity.Notification, error) {
	return nil, nil
}

func (notificationRepo *NotificationRepo) GetAllNotifications(userID int) ([]*entity.Notification, error) {
	return nil, nil
}

func notificationToInterfaces(notification *entity.Notification) []interface{} {
	notificationAsInterfaces := make([]interface{}, 6)
	notificationAsInterfaces[0] = uint(notification.NotificationID)
	notificationAsInterfaces[1] = uint(notification.UserID)
	notificationAsInterfaces[2] = notification.Category
	notificationAsInterfaces[3] = notification.Title
	notificationAsInterfaces[4] = notification.Text
	notificationAsInterfaces[5] = notification.IsRead
	return notificationAsInterfaces
}

func interfacesToNotification(interfaces []interface{}) *entity.Notification {
	notification := new(entity.Notification)
	notification.NotificationID = int(interfaces[0].(uint))
	notification.UserID = int(interfaces[1].(uint))
	notification.Category = interfaces[2].(string)
	notification.Title = interfaces[3].(string)
	notification.Text = interfaces[4].(string)
	notification.IsRead = interfaces[5].(bool)
	return notification
}
