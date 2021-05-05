package entity

type Message struct {
	MessageID      int
	AuthorID       int
	Text           string
	TimeOfCreation string
}

type MessageInput struct {
	MessageText string `json:"messageText"`
}

type MessageOutput struct {
	MessageID      int    `json:"ID"`
	ChatID         int    `json:"chatID"`
	AuthorID       int    `json:"authorID"`
	Text           string `json:"text"`
	TimeOfCreation string `json:"addingTime"`
}

// Fill from message fills MessageOutput using message and it's chatID
func (output *MessageOutput) FillFromMessage(message *Message, chatID int) {
	output.MessageID = message.MessageID
	output.ChatID = chatID
	output.AuthorID = message.AuthorID
	output.Text = message.Text
	output.TimeOfCreation = message.TimeOfCreation
}

type Chat struct {
	ChatID         int
	FirstUserID    int
	SecondUserID   int
	FirstUserRead  bool
	SecondUserRead bool
	Messages       map[int]Message
}

type ChatOutput struct {
	ChatID          int             `json:"ID"`
	TargetProfileID int             `json:"targetProfile"`
	Messages        []MessageOutput `json:"messages"`
	IsRead          bool            `json:"isRead"`
}

// FillFromChat fills ChatOutput from Chat
// isFirstUser is true if we intend to send chat to first user, false if to second
func (output *ChatOutput) FillFromChat(chat *Chat, isFirstUser bool) {
	output.ChatID = chat.ChatID
	output.Messages = make([]MessageOutput, 0)
	for _, message := range chat.Messages {
		var messageOutput MessageOutput
		messageOutput.FillFromMessage(&message, chat.ChatID)
		output.Messages = append(output.Messages, messageOutput)
	}

	switch isFirstUser {
	case true:
		output.TargetProfileID = chat.SecondUserID
		output.IsRead = chat.FirstUserRead
	case false:
		output.TargetProfileID = chat.FirstUserID
		output.IsRead = chat.SecondUserRead
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
	Type    key           `json:"type"`
	Message MessageOutput `json:"message"`
}
