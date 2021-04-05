package repository

import "pinterest/domain/entity"

type UserRepository interface {
	SaveUser(*entity.User) (*entity.User, map[string]string)
	GetUser(int) (*entity.User, error) // Get user by his ID
	GetUsers() ([]entity.User, error)  // Get all users
	GetUserByUsername(string) (*entity.User, map[string]string)
	CheckUserCredentials(string, string) (bool, error) // Check if passed username and password are correct
}
