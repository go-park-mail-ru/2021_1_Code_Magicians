package entity

type Board struct {
	BoardID     int    `json:"ID"`
	UserID      int    `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
