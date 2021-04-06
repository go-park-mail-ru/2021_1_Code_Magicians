package repository

import "pinterest/domain/entity"

type UserRepository interface {
	SaveUser(*entity.User) (int, error) // Returns saved user's ID
	DeleteUser(int) error
	GetUser(int) (*entity.User, error) // Get user by his ID
	GetUsers() ([]entity.User, error)  // Get all users
	GetUserByUsername(string) (*entity.User, error)
}
