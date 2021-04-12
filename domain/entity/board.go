package entity

type Board struct {
	BoardID     int    `json:"boardID"`
	UserID      int    `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
