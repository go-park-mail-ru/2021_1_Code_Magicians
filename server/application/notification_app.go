package application

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
	"strings"

	"text/template"
)

type NotificationApp struct {
	notificationRepo repository.NotificationRepositoryInterface
	userApp          UserAppInterface
	websocketApp     WebsocketAppInterface
}

func NewNotificationApp(notificationRepo repository.NotificationRepositoryInterface,
	userApp UserAppInterface, websocketApp WebsocketAppInterface) *NotificationApp {
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
	SendNotificationsToUsers(usersAndNotifications []entity.UserNotificationInfo) // Send notifications to users
	SendNotificationEmail(userID int, notificationID int,
		templateString string, templateStruct interface{},
		serverEmail string, serverPassword string) error // Send specified notification to specified user as an e-mail
	ReadNotification(userID int, notificationID int) error // Changes notification's status to "Read"
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

func (notificationApp *NotificationApp) SendNotificationsToUsers(usersAndNotifications []entity.UserNotificationInfo) {
	for _, pair := range usersAndNotifications {
		notificationApp.SendNotification(pair.UserID, pair.NotificationID)
	}
}

// SendMailTLS not use STARTTLS commond
func sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	tlsconfig := &tls.Config{ServerName: host}
	if err = validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err = validateLine(recp); err != nil {
			return err
		}
	}
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// validateLine checks to see if a line has CR or LF as per RFC 5321
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("a line must not contain CR or LF")
	}
	return nil
}

func (notificationApp *NotificationApp) SendNotificationEmail(userID int, notificationID int,
	templateString string, templateStruct interface{},
	serverEmail string, serverPassword string) error {
	user, err := notificationApp.userApp.GetUser(userID)
	if err != nil {
		return err
	}

	notification, err := notificationApp.GetNotification(userID, notificationID)
	if err != nil {
		return err
	}

	// Receiver email address.
	to := []string{
		user.Email,
	}

	// smtp server configuration.
	smtpHost := "smtp.mail.ru"
	smtpPort := "465"

	// Authentication.
	auth := smtp.PlainAuth("", serverEmail, serverPassword, smtpHost)

	t, err := template.New("EMail Template").Parse(templateString)
	if err != nil {
		return err
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s", notification.Title, mimeHeaders)))

	t.Execute(&body, templateStruct)

	// Sending email.
	err = sendMailTLS(smtpHost+":"+smtpPort, auth, serverEmail, to, body.Bytes())
	if err != nil {
		return err
	}

	return nil
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
