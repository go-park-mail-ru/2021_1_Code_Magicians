package repository

import "pinterest/domain/entity"

type UserRepository interface {
	CreateUser(*entity.User) (int, error)           // Create user, returns created user's ID
	SaveUser(*entity.User) error                    // Save changed user to database
	DeleteUser(int) error                           // Delete user with passed userID from database
	GetUser(int) (*entity.User, error)              // Get user by his ID
	GetUsers() ([]entity.User, error)               // Get all users
	GetUserByUsername(string) (*entity.User, error) // Get user by his username
}
