package persistance

import (
	"pinterest/domain/entity"

	"github.com/tarantool/go-tarantool"
)

type ChatRepo struct {
	tarantoolDB *tarantool.Connection
}

func NewChatRepository(tarantoolDB *tarantool.Connection) *ChatRepo {
	return &ChatRepo{tarantoolDB}
}

func (chatRepo *ChatRepo) CreateChat(firstUserID int, secondUserID int) (int, error) {
	_, err := chatRepo.GetChatIDByUsers(firstUserID, secondUserID)
	if err != entity.ChatNotFoundError {
		if err == nil {
			return -1, entity.ChatAlreadyExistsError
		}
		return -1, err
	}

	resp, err := chatRepo.tarantoolDB.Insert("chats", []interface{}{nil, firstUserID, secondUserID, false, false})
	if err != nil {
		return -1, err
	}

	newChatID := int(resp.Tuples()[0][0].(uint64))
	return newChatID, nil
}

func (chatRepo *ChatRepo) GetChat(chatID int) (*entity.Chat, error) {
	resp, err := chatRepo.tarantoolDB.Select("chats", "primary", 0, 1, tarantool.IterEq, []interface{}{uint(chatID)})
	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.ChatNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.ChatNotFoundError
	}

	return interfacesToChat(resp.Tuples()[0]), nil
}

func (chatRepo *ChatRepo) getAllChatsByIndex(userID int, index string) ([]*entity.Chat, error) {
	const MaxUint32 = ^uint32(0) // So that upper limit for select is practically "infinity"
	resp, err := chatRepo.tarantoolDB.Select("chats", index, 0, MaxUint32, tarantool.IterEq, []interface{}{uint(userID)})
	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.ChatsNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) == 0 {
		return nil, entity.ChatsNotFoundError
	}

	chats := make([]*entity.Chat, 0, len(resp.Tuples()))

	for _, tuple := range resp.Tuples() {
		chats = append(chats, interfacesToChat(tuple))
	}

	return chats, nil
}

func (chatRepo *ChatRepo) GetAllChats(userID int) ([]*entity.Chat, error) {
	firstChats, err := chatRepo.getAllChatsByIndex(userID, "by_first_user")
	if err != nil {
		if err != entity.ChatsNotFoundError {
			return nil, err
		}
		firstChats = make([]*entity.Chat, 0)
	}

	secondChats, err := chatRepo.getAllChatsByIndex(userID, "by_second_user")
	if err != nil {
		if err != entity.ChatsNotFoundError {
			return nil, err
		}
		secondChats = make([]*entity.Chat, 0)
	}

	firstChats = append(firstChats, secondChats...)
	if len(firstChats) == 0 {
		return nil, entity.ChatsNotFoundError
	}

	return firstChats, nil
}

func (chatRepo *ChatRepo) SaveChat(chat *entity.Chat) error {
	updateCommand := []interface{}{[]interface{}{"=", 3, chat.FirstUserRead}, []interface{}{"=", 4, chat.SecondUserRead}}
	_, err := chatRepo.tarantoolDB.Update("chats", "primary", []interface{}{uint(chat.ChatID)}, updateCommand)

	return err
}

func (chatRepo *ChatRepo) GetChatIDByUsers(firstUserID int, secondUserID int) (int, error) {
	resp, err := chatRepo.tarantoolDB.Select("chats", "secondary", 0, 1, tarantool.IterEq, []interface{}{uint(firstUserID), uint(secondUserID)})
	if resp == nil {
		return -1, err
	}

	if len(resp.Tuples()) != 1 {
		resp, err = chatRepo.tarantoolDB.Select("chats", "secondary", 0, 1, tarantool.IterEq, []interface{}{uint(secondUserID), uint(firstUserID)})
		if resp == nil {
			return -1, err
		}

		if len(resp.Tuples()) != 1 {
			return -1, entity.ChatNotFoundError
		}
	}

	newChatID := int(resp.Tuples()[0][0].(uint64))

	return newChatID, nil
}

func (chatRepo *ChatRepo) AddMessage(message *entity.Message) (int, error) {
	messageInterfaces := messageToInterfaces(message)
	messageInterfaces[0] = nil
	resp, err := chatRepo.tarantoolDB.Insert("messages", messageInterfaces)
	if err != nil {
		if resp == nil {
			return -1, err
		}

		return -1, err
	}

	if len(resp.Tuples()) != 1 {
		return -1, entity.MessageAddingError
	}

	newMessageID := int(resp.Tuples()[0][0].(uint64))

	return newMessageID, nil
}

func (chatRepo *ChatRepo) GetMessage(messageID int) (*entity.Message, error) {
	resp, err := chatRepo.tarantoolDB.Select("messages", "primary", 0, 1, tarantool.IterEq, []interface{}{uint(messageID)})
	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.MessageNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) != 1 {
		return nil, entity.MessageNotFoundError
	}

	return interfacesToMessage(resp.Tuples()[0]), nil
}

func (chatRepo *ChatRepo) GetMessages(chatID int) ([]*entity.Message, error) {
	const MaxUint32 = ^uint32(0) // So that upper limit for select is practically "infinity"
	resp, err := chatRepo.tarantoolDB.Select("messages", "secondary", 0, MaxUint32, tarantool.IterEq, []interface{}{uint(chatID)})
	if err != nil {
		if resp == nil {
			return nil, err
		}

		switch resp.Code {
		case tarantool.ErrTupleNotFound:
			return nil, entity.MessagesNotFoundError
		default:
			return nil, err
		}
	}

	if len(resp.Tuples()) == 0 {
		return nil, entity.MessagesNotFoundError
	}

	messages := make([]*entity.Message, 0, len(resp.Tuples()))

	for _, tuple := range resp.Tuples() {
		messages = append(messages, interfacesToMessage(tuple))
	}

	return messages, nil
}

func chatToInterfaces(chat *entity.Chat) []interface{} {
	chatAsInterfaces := make([]interface{}, 5)
	chatAsInterfaces[0] = uint(chat.ChatID)
	chatAsInterfaces[1] = uint(chat.FirstUserID)
	chatAsInterfaces[2] = uint(chat.SecondUserID)
	chatAsInterfaces[3] = chat.FirstUserRead
	chatAsInterfaces[4] = chat.SecondUserRead
	return chatAsInterfaces
}

func interfacesToChat(interfaces []interface{}) *entity.Chat {
	chat := new(entity.Chat)
	chat.ChatID = int(interfaces[0].(uint64))
	chat.FirstUserID = int(interfaces[1].(uint64))
	chat.SecondUserID = int(interfaces[2].(uint64))
	chat.FirstUserRead = interfaces[3].(bool)
	chat.SecondUserRead = interfaces[4].(bool)
	return chat
}

func messageToInterfaces(message *entity.Message) []interface{} {
	messageAsInterfaces := make([]interface{}, 5)
	messageAsInterfaces[0] = uint(message.MessageID)
	messageAsInterfaces[1] = uint(message.ChatID)
	messageAsInterfaces[2] = uint(message.AuthorID)
	messageAsInterfaces[3] = message.Text
	messageAsInterfaces[4] = message.TimeOfCreation
	return messageAsInterfaces
}

func interfacesToMessage(interfaces []interface{}) *entity.Message {
	message := new(entity.Message)
	message.MessageID = int(interfaces[0].(uint64))
	message.ChatID = int(interfaces[1].(uint64))
	message.AuthorID = int(interfaces[2].(uint64))
	message.Text = interfaces[3].(string)
	message.TimeOfCreation = interfaces[4].(string)
	return message
}
