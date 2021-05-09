package repository

import "pinterest/domain/entity"

type UserRepository interface {
	CreateUser(user *entity.User) (int, error)                    // Create user, returns created user's ID
	SaveUser(user *entity.User) error                             // Save changed user to database
	DeleteUser(userID int) error                                  // Delete user with passed userID from database
	GetUser(userID int) (*entity.User, error)                     // Get user by his ID
	GetUsers() ([]entity.User, error)                             // Get all users
	GetUserByUsername(username string) (*entity.User, error)      // Get user by his username
	Follow(followerId int, followedID int) error                  // Make first user follow second
	Unfollow(followerID int, followedID int) error                // Make first user unfollow second
	CheckIfFollowed(followerID int, followedID int) (bool, error) // Check if first user follows second
	SearchUsers(keywords string) ([]entity.User, error)           // Get all users by passed keywords
}
