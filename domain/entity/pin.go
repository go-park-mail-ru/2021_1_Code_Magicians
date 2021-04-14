package entity

type Pin struct {
	PinId       int    `json:"id"`
	UserID      int    `json:"userID"`
	Title       string `json:"title"`
	ImageLink   string `json:"pinImage"`
	Description string `json:"description"`
}
