package repository

import "pinterest/domain/entity"

type ChatRepositoryInterface interface {
	CreateChat(firstUserID int, secondUserID int) (int, error)       // Create chat between first and second user (errors if chat exists already)
	GetChat(chatID int) (*entity.Chat, error)                        // Get chat by it's ID
	GetAllChats(userID int) ([]*entity.Chat, error)                  // Get all chats that specified user is in
	SaveChat(chat *entity.Chat) error                                // Save chat in database, replacing one with same chatID
	GetChatIDByUsers(firstUserID int, secondUserID int) (int, error) // Find chat between specified users
	AddMessage(message *entity.Message) (int, error)                 // Add message to database
	GetMessage(messageID int) (*entity.Message, error)               // Get message by specified ID
	GetMessages(chatID int) ([]*entity.Message, error)               // Return all messages from specified chat
}
