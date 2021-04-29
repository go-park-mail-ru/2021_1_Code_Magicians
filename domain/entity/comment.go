package entity

type Comment struct {
	UserID     int    `json:"userID"`
	PinID      int    `json:"pinID"`
	PinComment string `json:"text"`
}

type CommentTextOutput struct {
	Text string `json:"text"`
}

type CommentsOutput struct {
	Comments []Comment `json:"comments"`
}