package entity

type InitialMessage struct {
	UserID    int    `json:"userID"`
	CSRFToken string `json:"CSRFToken"`
}
