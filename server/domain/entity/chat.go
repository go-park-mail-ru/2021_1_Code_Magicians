package entity

type Message struct {
	MessageID      int    `json:"ID"`
	ChatID         int    `json:"chatID"`
	AuthorID       int    `json:"authorID"`
	Text           string `json:"text"`
	TimeOfCreation string `json:"addingTime"`
}

type MessageInput struct {
	MessageText string `json:"messageText"`
}

type Chat struct {
	ChatID         int
	FirstUserID    int
	SecondUserID   int
	FirstUserRead  bool
	SecondUserRead bool
}

type ChatOutput struct {
	ChatID        int        `json:"ID"`
	TargetProfile UserOutput `json:"targetProfile"`
	Messages      []Message  `json:"messages"`
	IsRead        bool       `json:"isRead"`
}

// FillFromChat fills ChatOutput from Chat
func (output *ChatOutput) FillFromChat(chat *Chat, target *User, messages []*Message) {
	output.ChatID = chat.ChatID
	output.Messages = make([]Message, 0, len(messages))
	for _, message := range messages {
		output.Messages = append(output.Messages, *message)
	}

	var targetProfileOutput UserOutput
	targetProfileOutput.FillFromUser(target)
	targetProfileOutput.Email = ""
	output.TargetProfile = targetProfileOutput

	switch chat.FirstUserID == target.UserID {
	case true:
		output.IsRead = chat.SecondUserRead
	case false:
		output.IsRead = chat.FirstUserRead
	}
}

type AllChatsOutput struct {
	Type  key          `json:"type"`
	Chats []ChatOutput `json:"allChats"`
}

type OneChatOutput struct {
	Type key        `json:"type"`
	Chat ChatOutput `json:"chat"`
}

type OneMessageOutput struct {
	Type    key     `json:"type"`
	Message Message `json:"message"`
}
