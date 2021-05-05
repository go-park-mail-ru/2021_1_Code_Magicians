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
	ChatID        int             `json:"ID"`
	TargetProfile UserOutput      `json:"targetProfile"`
	Messages      []MessageOutput `json:"messages"`
	IsRead        bool            `json:"isRead"`
}

// FillFromChat fills ChatOutput from Chat
func (output *ChatOutput) FillFromChat(chat *Chat, target *User) {
	output.ChatID = chat.ChatID
	output.Messages = make([]MessageOutput, 0)
	for _, message := range chat.Messages {
		var messageOutput MessageOutput
		messageOutput.FillFromMessage(&message, chat.ChatID)
		output.Messages = append(output.Messages, messageOutput)
	}

	var targetProfileOutput UserOutput
	targetProfileOutput.FillFromUser(target)
	targetProfileOutput.Email = ""
	output.TargetProfile = targetProfileOutput

	switch chat.FirstUserID == target.UserID {
	case true:
		output.IsRead = chat.FirstUserRead
	case false:
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
