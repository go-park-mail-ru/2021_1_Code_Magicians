package entity

type Board struct {
	BoardID       int    `json:"ID"`
	UserID        int    `json:"userID"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageLink     string `json:"avatarLink"`
	ImageHeight   int    `json:"avatarHeight"`
	ImageWidth    int    `json:"avatarWidth"`
	ImageAvgColor string `json:"avatarAvgColor"`
}

type BoardsOutput struct {
	Boards []Board `json:"boards"`
}

type BoardID struct {
	BoardID int `json:"ID"`
}
