package entity

type Notification struct {
	UserID         int    `json:"-"`
	NotificationID int    `json:"ID"`
	Title          string `json:"title"`
	Category       string `json:"category"`
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
