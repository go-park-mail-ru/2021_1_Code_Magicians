package entity

type Pin struct {
	PinID         int    `json:"ID"`
	UserID        int    `json:"userID"`
	BoardID       int    `json:"boardID,omitempty"`
	Title         string `json:"title"`
	ImageLink     string `json:"imageLink"`
	ImageHeight   int    `json:"imageHeight"`
	ImageWidth    int    `json:"imageWidth"`
	ImageAvgColor string `json:"imageAvgColor"`
	Description   string `json:"description"`
}

type PinsOutput struct {
	Pins []Pin `json:"pins"`
}

type PinID struct {
	PinID int `json:"ID"`
}
