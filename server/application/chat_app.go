package application

import (
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
