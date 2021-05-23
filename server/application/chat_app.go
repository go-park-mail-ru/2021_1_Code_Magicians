package application

import (
	"encoding/json"
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type ChatApp struct {
	chatRepo     repository.ChatRepositoryInterface
	userApp      UserAppInterface
	websocketApp WebsocketAppInterface
}

func NewChatApp(chatRepo repository.ChatRepositoryInterface,
	userApp UserAppInterface, websocketApp WebsocketAppInterface) *ChatApp {
	return &ChatApp{
		chatRepo:     chatRepo,
		userApp:      userApp,
		websocketApp: websocketApp,
	}
}

type ChatAppInterface interface {
	CreateChat(firstUserID int, secondUserID int) (int, error)       // Create chat between first and second user (errors if chat exists already)
	GetChatIDByUsers(firstUserID int, secondUserID int) (int, error) // Find chat between specified users
	AddMessage(message *entity.Message) (int, error)                 // Add message (author has to be in message's chat)
	SendMessage(chatID int, messageID int, userID int) error         // Send specified message from specified chat to user (who must be in said chat)
	SendChat(userID int, chatId int) error                           // Send entire specified chat to specified user (who  must be in said chat)
	SendAllChats(userID int) error                                   // Send all chats of specified user to them
	ReadChat(chatID int, userID int) error                           // Mark specified chat as "Read" for specified user
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

	return chatApp.chatRepo.CreateChat(firstUserID, secondUserID)
}

func (chatApp *ChatApp) GetChatIDByUsers(firstUserID int, secondUserID int) (int, error) {
	return chatApp.chatRepo.GetChatIDByUsers(firstUserID, secondUserID)
}

func (chatApp *ChatApp) AddMessage(message *entity.Message) (int, error) {
	chat, err := chatApp.chatRepo.GetChat(message.ChatID)
	if err != nil {
		return -1, err
	}

	switch {
	case chat.FirstUserID == message.AuthorID:
		chat.FirstUserRead = true
		chat.SecondUserRead = false
	case chat.SecondUserID == message.AuthorID:
		chat.SecondUserRead = true
		chat.FirstUserRead = false
	default:
		return -1, entity.UserNotInChatError
	}

	messageID, err := chatApp.chatRepo.AddMessage(message)
	if err != nil {
		return -1, err
	}

	chatApp.chatRepo.SaveChat(chat)
	return messageID, nil
}

func (chatApp *ChatApp) SendMessage(chatID int, messageID int, userID int) error {
	chat, err := chatApp.chatRepo.GetChat(chatID)
	if err != nil {
		return err
	}

	if chat.FirstUserID != userID && chat.SecondUserID != userID {
		return entity.UserNotInChatError
	}

	message, err := chatApp.chatRepo.GetMessage(messageID)
	if err != nil {
		return err
	}

	messageOutputMsg := entity.OneMessageOutput{Type: entity.OneMessageTypeKey, Message: *message} // TODO: fix naming

	result, err := json.Marshal(messageOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}

func (chatApp *ChatApp) SendChat(chatID int, userID int) error {
	chat, err := chatApp.chatRepo.GetChat(chatID)
	if err != nil {
		return err
	}

	if chat.FirstUserID != userID && chat.SecondUserID != userID {
		return entity.UserNotInChatError
	}

	var target *entity.User
	switch {
	case chat.FirstUserID == userID:
		target, err = chatApp.userApp.GetUser(chat.SecondUserID)
		if err != nil {
			return entity.UserNotFoundError
		}
	case chat.SecondUserID == userID:
		target, err = chatApp.userApp.GetUser(chat.FirstUserID)
		if err != nil {
			return entity.UserNotFoundError
		}
	default:
		return entity.UserNotInChatError
	}

	messages, err := chatApp.chatRepo.GetMessages(chatID)
	if err != nil {
		return err
	}

	var chatOutput entity.ChatOutput
	chatOutput.FillFromChat(chat, target, messages)

	chatOutputMsg := entity.OneChatOutput{Type: entity.OneChatTypeKey, Chat: chatOutput}

	result, err := json.Marshal(chatOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}

func (chatApp *ChatApp) SendAllChats(userID int) error { // O(n) now, will be log(n) when i'll add actual database
	target, err := chatApp.userApp.GetUser(userID)
	if err != nil {
		return err
	}

	chats, err := chatApp.chatRepo.GetAllChats(userID)
	if err != nil {
		if err != entity.ChatsNotFoundError {
			return err
		}
		chats = make([]*entity.Chat, 0)
	}

	chatOutputs := make([]entity.ChatOutput, 0, len(chats))
	for _, chat := range chats {
		messages, err := chatApp.chatRepo.GetMessages(chat.ChatID)
		if err != nil {
			if err != entity.MessagesNotFoundError {
				return err
			}
			messages = make([]*entity.Message, 0)
		}

		var chatOutput entity.ChatOutput
		chatOutput.FillFromChat(chat, target, messages)
		chatOutputs = append(chatOutputs, chatOutput)
	}

	chatsOutputMsg := entity.AllChatsOutput{Type: entity.AllChatsTypeKey, Chats: chatOutputs}

	result, err := json.Marshal(chatsOutputMsg)
	if err != nil {
		return entity.JsonMarshallError
	}

	err = chatApp.websocketApp.SendMessage(userID, result)

	return err
}

func (chatApp *ChatApp) ReadChat(chatID int, userID int) error {
	chat, err := chatApp.chatRepo.GetChat(chatID)
	if err != nil {
		return err
	}

	switch {
	case chat.FirstUserID == userID:
		chat.FirstUserRead = true
	case chat.SecondUserID == userID:
		chat.SecondUserRead = true
	default:
		return entity.UserNotInChatError
	}

	return chatApp.chatRepo.SaveChat(chat)
}
