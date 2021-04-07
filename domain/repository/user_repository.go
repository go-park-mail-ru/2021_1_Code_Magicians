package repository

import "pinterest/domain/entity"

type UserRepository interface {
	CreateUser(*entity.User) (int, error) // Returns saved user's ID
	SaveUser(*entity.User) error          // Saves changes to user
	DeleteUser(int) error
	GetUser(int) (*entity.User, error) // Get user by his ID
	GetUsers() ([]entity.User, error)  // Get all users
	GetUserByUsername(string) (*entity.User, error)
}
