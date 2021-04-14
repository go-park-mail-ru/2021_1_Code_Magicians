package entity

type Pin struct {
	PinId       int    `json:"id"`
	UserID      int    `json:"userID"`
	BoardID		int	   `json:"boardID,omitempty"`
	Title       string `json:"title"`
	ImageLink   string `json:"pinImage"`
	Description string `json:"description"`
}
