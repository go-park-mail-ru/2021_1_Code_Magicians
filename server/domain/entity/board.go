package entity

type Board struct {
	BoardID     int    `json:"ID"`
	UserID      int    `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageLInk   string `json:"avatarLink"`
}

type BoardInfo struct {
	BoardID     int    `json:"ID"`
	UserID      int    `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageLink   string `json:"avatarLink,omitempty"`
	//Pins        []Pin  `json:"pins"`
}

type BoardsOutput struct {
	Boards []Board `json:"boards"`
}

type BoardID struct {
	BoardID int `json:"ID"`
}
