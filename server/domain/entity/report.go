package entity

type Report struct {
	ReportID    int    `json:"reportID"`
	PinID       int    `json:"pinID"`
	SenderID    int    `json:"senderID"`
	Description string `json:"description"`
}

type ReportID struct {
	ReportID int `json:"reportID"`
}
