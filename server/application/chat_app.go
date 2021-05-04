package application

import (
	"encoding/json"
	"pinterest/domain/entity"
	"sync"
)

type ChatApp struct {
	chats         map[int]entity.Chat
	lastChatID    int
	lastMessageID int
	mu            sync.Mutex
	userApp       UserAppInterface
	websocketApp  WebsocketAppInterface
}

func NewChatApp(userApp UserAppInterface, websocketApp WebsocketAppInterface) *ChatApp {
	return &ChatApp{
		chats:        make(map[int]entity.Chat),
		userApp:      userApp,
		websocketApp: websocketApp,
	}
}

type ChatAppInterface interface {
	CreateChat(firstUserID int, secondUserID int) (int, error)   // Create chat between first and second user (errors if chat exists already)
	AddMessage(chatId int, message *entity.Message) (int, error) // Add message to specified chat (author has to be in said chat)
	SendMessage(chatID int, messageID int, userID int) error     // Send specified message from specified chat to user (who must be in said chat)
	SendChat(userID int, chatId int) error                       // Send entire specified chat to specified user (who  must be in said chat)
	SendAllChats(userID int) error                               // Send all chats of specified user to them
}

func (chatApp *ChatApp) getChatByUsers(firstUserID int, SecondUserID int) (*entity.Chat, error) {
	chatApp.mu.Lock()
	defer chatApp.mu.Unlock()

	for _, chat := range chatApp.chats {
		if chat.FirstUserID == firstUserID && chat.SecondUserID == SecondUserID {
			return &chat, nil
		}
		if chat.FirstUserID == SecondUserID && chat.SecondUserID == firstUserID { // User's actual order does not matter
			return &chat, nil
		}
	}

	return nil, entity.ChatNotFoundError
}

func (chatApp *ChatApp) getChatByID(chatID int) (*entity.Chat, error) {
	chatApp.mu.Lock()
	defer chatApp.mu.Unlock()

	chat, found := chatApp.chats[chatID]
	if !found {
		return nil, entity.ChatNotFoundError
	}

	return &chat, nil
}

func (chatApp *ChatApp) CreateChat(firstUserID int, secondUserID int) (int, error) {
	_, err := chatApp.userApp.GetUser(firstUserID)
	if err != nil {
		return -1, entity.UserNotFoundError
	}
	_, err = chatApp.userApp.GetUser(secondUserID)
	if err != nil {
		return -1, entity.UserNotFoundError
	}

	_, err = chatApp.getChatByUsers(firstUserID, secondUserID)
	if err == nil {
		return -1, entity.ChatAlreadyExistsError
	}

	chatApp.mu.Lock()
	chat := entity.Chat{
		ChatID:         chatApp.lastChatID,
		FirstUserID:    firstUserID,
		SecondUserID:   secondUserID,
		FirstUserRead:  false,
		SecondUserRead: false,
		Messages:       make(map[int]entity.Message),
	}
	chatApp.chats[chat.ChatID] = chat
	chatApp.lastChatID++
	chatApp.mu.Unlock()

	return chat.ChatID, nil
}

func (chatApp *ChatApp) AddMessage(chatID int, message *entity.Message) (int, error) {
	chat, err := chatApp.getChatByID(chatID)
	if err != nil {
		return -1, err
	}

	if chat.FirstUserID != message.AuthorID && chat.SecondUserID != message.AuthorID {
		return -1, entity.UserNotInChatError
	}

	chatApp.mu.Lock()
	message.MessageID = chatApp.lastMessageID
	chatApp.lastMessageID++

	chat.Messages[message.MessageID] = *message
	chatApp.chats[chatID] = *chat
	chatApp.mu.Unlock()

	return message.MessageID, nil
}

func (chatApp *ChatApp) SendMessage(chatID int, messageID int, userID int) error {
	chat, err := chatApp.getChatByID(chatID)
	if err != nil {
		return err
	}

	if chat.FirstUserID != userID && chat.SecondUserID != userID {
		return entity.UserNotInChatError
	}

	message, found := chat.Messages[messageID]
	if !found {
		return entity.MessageNotFoundError
	}

	var messageOutput entity.MessageOutput
	messageOutput.FillFromMessage(&message, chatID)

	messageOutputMsg := entity.OneMessageOutput{Type: entity.OneMessageTypeKey, Message: messageOutput} // TODO: fix naming

	result, err := json.Marshal(messageOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}

func (chatApp *ChatApp) SendChat(chatID int, userID int) error {
	chat, err := chatApp.getChatByID(chatID)
	if err != nil {
		return err
	}

	if chat.FirstUserID != userID && chat.SecondUserID != userID {
		return entity.UserNotInChatError
	}

	var chatOutput entity.ChatOutput
	chatOutput.FillFromChat(chat, chat.FirstUserID == userID)

	chatOutputMsg := entity.OneChatOutput{Type: entity.OneChatTypeKey, Chat: chatOutput}

	result, err := json.Marshal(chatOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}

func (chatApp *ChatApp) SendAllChats(userID int) error { // O(n) now, will be log(n) when i'll add actual database
	_, err := chatApp.userApp.GetUser(userID)
	if err != nil {
		return err
	}

	chatOutputs := make([]entity.ChatOutput, 0)

	chatApp.mu.Lock()
	for _, chat := range chatApp.chats {
		if chat.FirstUserID == userID || chat.SecondUserID == userID {
			var chatOutput entity.ChatOutput
			chatOutput.FillFromChat(&chat, chat.FirstUserID == userID) // TODO: also set isRead states
			chatOutputs = append(chatOutputs, chatOutput)
		}
	}
	chatApp.mu.Unlock()

	chatsOutputMsg := entity.AllChatsOutput{Type: entity.OneChatTypeKey, Chats: chatOutputs}

	result, err := json.Marshal(chatsOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}
