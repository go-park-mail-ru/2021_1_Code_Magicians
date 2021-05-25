package persistance

import (
	"fmt"
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
	notificationAsInterface[0] = nil // Because we don't know notification's ID
	resp, err := notificationRepo.tarantoolDB.Insert("notifications", notificationAsInterface)
	if err != nil {
		return -1, err
	}

	if len(resp.Tuples()) != 1 {
		return -1, fmt.Errorf("Could not add notification")
	}

	newNotificationID := int(resp.Tuples()[0][0].(uint64))
	return newNotificationID, nil
}

func (notificationRepo *NotificationRepo) RemoveNotification(notificationID int) error {
	_, err := notificationRepo.tarantoolDB.Delete("notifications", "primary", []interface{}{uint(notificationID)})
	return err
}

func (notificationRepo *NotificationRepo) EditNotification(notification *entity.Notification) error {
	updateCommand := []interface{}{[]interface{}{"=", 1, uint(notification.UserID)}, []interface{}{"=", 2, notification.Category},
		[]interface{}{"=", 3, notification.Title}, []interface{}{"=", 4, notification.Text}, []interface{}{"=", 5, notification.IsRead}}
	_, err := notificationRepo.tarantoolDB.Update(
		"notifications", "primary", []interface{}{uint(notification.NotificationID)}, updateCommand,
	)
	if err != nil {
		return err
	}

	return nil
}

func (notificationRepo *NotificationRepo) GetNotification(notificationID int) (*entity.Notification, error) {
	resp, err := notificationRepo.tarantoolDB.Select("notifications", "primary", 0, 1, tarantool.IterEq, []interface{}{uint(notificationID)})

	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.NotificationNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.NotificationNotFoundError
	}

	notification := interfacesToNotification(resp.Tuples()[0])

	return notification, nil
}

func (notificationRepo *NotificationRepo) GetAllNotifications(userID int) ([]*entity.Notification, error) {
	const MaxUint32 = ^uint32(0) // So that upper limit for select is practically "infinity"
	resp, err := notificationRepo.tarantoolDB.Select("notifications", "secondary", 0, MaxUint32, tarantool.IterEq, []interface{}{uint(userID)})

	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.NotificationNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) == 0 {
		return nil, entity.NotificationsNotFoundError
	}

	notifications := make([]*entity.Notification, 0, len(resp.Tuples()))

	for _, tuple := range resp.Tuples() {
		notifications = append(notifications, interfacesToNotification(tuple))
	}

	return notifications, nil
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
	notification.NotificationID = int(interfaces[0].(uint64))
	notification.UserID = int(interfaces[1].(uint64))
	notification.Category = interfaces[2].(string)
	notification.Title = interfaces[3].(string)
	notification.Text = interfaces[4].(string)
	notification.IsRead = interfaces[5].(bool)
	return notification
}
