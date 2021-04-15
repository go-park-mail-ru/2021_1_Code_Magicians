package entity

type Pin struct {
	PinId       int    `json:"ID"`
	UserID      int    `json:"userID"`
	BoardID		int	   `json:"boardID,omitempty"`
	Title       string `json:"title"`
	ImageLink   string `json:"imageLink"`
	Description string `json:"description"`
}
