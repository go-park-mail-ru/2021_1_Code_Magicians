package entity

type Notification struct {
	NotificationID int    `json:"ID"`
	UserID         int    `json:"-"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Text           string `json:"text"`
	IsRead         bool   `json:"isRead"`
}

type AllNotificationsOutput struct {
	Type          key            `json:"type"`
	Notifications []Notification `json:"allNotifications"`
}

type OneNotificationOutput struct {
	Type         key          `json:"type"`
	Notification Notification `json:"notification"`
}
