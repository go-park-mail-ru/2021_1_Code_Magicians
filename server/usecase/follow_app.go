package usecase

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type FollowApp struct {
	userRepository repository.UserRepository
}

func NewFollowApp(userRepository repository.UserRepository) *FollowApp {
	return &FollowApp{userRepository}
}

type FollowAppInterface interface {
	Follow(followerID int, followedID int) error                  // Make first user follow second
	Unfollow(followerID int, followedID int) error                // Make first user unfollow second
	CheckIfFollowed(followerID int, followedID int) (bool, error) // Check if first user follows second. Err != nil if those users are the same
	GetAllFollowers(followedID int) ([]entity.User, error)        // Get everyone who follows specified user
	GetAllFollowed(followerID int) ([]entity.User, error)         // Get everyone who is followed by specified user
}

func (followApp *FollowApp) Follow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return followApp.userRepository.Follow(followerID, followedID)
}

func (followApp *FollowApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}
	return followApp.userRepository.Unfollow(followerID, followedID)
}

func (followApp *FollowApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.SelfFollowError
	}
	return followApp.userRepository.CheckIfFollowed(followerID, followedID)
}

func (followApp *FollowApp) GetAllFollowers(followedID int) ([]entity.User, error) {
	_, err := followApp.userRepository.GetUser(followedID)
	if err != nil {
		return nil, err
	}
	return followApp.userRepository.GetAllFollowers(followedID)
}

func (followApp *FollowApp) GetAllFollowed(followerID int) ([]entity.User, error) {
	_, err := followApp.userRepository.GetUser(followerID)
	if err != nil {
		return nil, err
	}
	return followApp.userRepository.GetAllFollowed(followerID)
}
