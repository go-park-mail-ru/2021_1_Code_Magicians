package usecase

import (
	"pinterest/domain/entity"
	"pinterest/domain/repository"
)

type FollowApp struct {
	userRepository repository.UserRepository
	pinApp         PinAppInterface
}

func NewFollowApp(userRepository repository.UserRepository, pinApp PinAppInterface) *FollowApp {
	return &FollowApp{
		userRepository: userRepository,
		pinApp:         pinApp,
	}
}

type FollowAppInterface interface {
	Follow(followerID int, followedID int) error                  // Make first user follow second
	Unfollow(followerID int, followedID int) error                // Make first user unfollow second
	CheckIfFollowed(followerID int, followedID int) (bool, error) // Check if first user follows second. Err != nil if those users are the same
	GetAllFollowers(followedID int) ([]entity.User, error)        // Get everyone who follows specified user
	GetAllFollowed(followerID int) ([]entity.User, error)         // Get everyone who is followed by specified user
	GetPinsOfFollowedUsers(userID int) ([]entity.Pin, error)      // Get all pins belonging to users that user follows
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

func (followApp *FollowApp) GetPinsOfFollowedUsers(userID int) ([]entity.Pin, error) {
	followedUsers, err := followApp.GetAllFollowed(userID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int, 0, len(followedUsers))
	for _, user := range followedUsers {
		userIDs = append(userIDs, user.UserID)
	}

	return followApp.pinApp.GetPinsOfUsers(userIDs)
}
