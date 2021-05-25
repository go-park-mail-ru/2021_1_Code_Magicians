package entity

import "time"

type Pin struct {
	PinID         int       `json:"ID"`
	UserID        int       `json:"userID"`
	BoardID       int       `json:"boardID"`
	Title         string    `json:"title"`
	ImageLink     string    `json:"imageLink"`
	ImageHeight   int       `json:"imageHeight"`
	ImageWidth    int       `json:"imageWidth"`
	ImageAvgColor string    `json:"imageAvgColor"`
	Description   string    `json:"description"`
	CreationDate  time.Time `json:"creationDate"`
}

type PinOutput struct {
	PinID         int    `json:"ID"`
	UserID        int    `json:"userID"`
	BoardID       int    `json:"boardID,omitempty"`
	Title         string `json:"title"`
	ImageLink     string `json:"imageLink"`
	ImageHeight   int    `json:"imageHeight"`
	ImageWidth    int    `json:"imageWidth"`
	ImageAvgColor string `json:"imageAvgColor"`
	Description   string `json:"description"`
	CreationDate  string `json:"creationDate"`
}

type PinsListOutput struct {
	Pins []PinOutput `json:"pins"`
}

type PinID struct {
	PinID int `json:"ID"`
}

func (pinOutput *PinOutput) FillFromPin(pin *Pin) {
	pinOutput.PinID = pin.PinID
	pinOutput.UserID = pin.UserID
	pinOutput.BoardID = pin.BoardID
	pinOutput.Title = pin.Title
	pinOutput.ImageLink = pin.ImageLink
	pinOutput.ImageHeight = pin.ImageHeight
	pinOutput.ImageWidth = pin.ImageWidth
	pinOutput.ImageAvgColor = pin.ImageAvgColor
	pinOutput.Description = pin.Description
	pinOutput.CreationDate = pin.CreationDate.String()
}
