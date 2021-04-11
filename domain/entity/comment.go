package entity

type Comment struct {
	PinComment string `json:"text"`
	UserID     int    `json:"userID"`
	PinID      int    `json:"pinID"`
}
