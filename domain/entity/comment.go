package entity

type Comment struct {
	UserID     int    `json:"userID"`
	PinID      int    `json:"pinID"`
	PinComment string `json:"text"`
}
