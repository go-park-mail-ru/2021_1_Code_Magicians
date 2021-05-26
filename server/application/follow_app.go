package application

import (
	"context"
	"pinterest/domain/entity"
	grpcUser "pinterest/services/user/proto"
	"strings"
)

type FollowApp struct {
	grpcClient grpcUser.UserClient
	pinApp     PinAppInterface
}

func NewFollowApp(grpcClient grpcUser.UserClient, pinApp PinAppInterface) *FollowApp {
	return &FollowApp{
		grpcClient: grpcClient,
		pinApp:     pinApp,
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

	_, err := followApp.grpcClient.Follow(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.FollowAlreadyExistsError.Error()):
			return entity.FollowAlreadyExistsError
		case strings.Contains(err.Error(), entity.UserNotFoundError.Error()):
			return entity.UserNotFoundError
		case strings.Contains(err.Error(), entity.FollowCountUpdateError.Error()):
			return entity.FollowCountUpdateError
		}
		return err
	}

	return nil
}

func (followApp *FollowApp) Unfollow(followerID int, followedID int) error {
	if followerID == followedID {
		return entity.SelfFollowError
	}

	_, err := followApp.grpcClient.Unfollow(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.FollowNotFoundError.Error()):
			return entity.FollowNotFoundError
		case strings.Contains(err.Error(), entity.FollowCountUpdateError.Error()):
			return entity.FollowCountUpdateError
		}
		return err
	}

	return nil
}

func (followApp *FollowApp) CheckIfFollowed(followerID int, followedID int) (bool, error) {
	if followerID == followedID {
		return false, entity.SelfFollowError
	}

	isFollowed, err := followApp.grpcClient.CheckIfFollowed(context.Background(), &grpcUser.Follows{FollowedID: int64(followedID), FollowerID: int64(followerID)})
	return isFollowed.IsFollowed, err
}

func (followApp *FollowApp) GetAllFollowers(followedID int) ([]entity.User, error) {
	_, err := followApp.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(followedID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return nil, entity.UserNotFoundError
		}
		return nil, err
	}

	followersList, err := followApp.grpcClient.GetAllFollowers(context.Background(), &grpcUser.UserID{Uid: int64(followedID)})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.UsersNotFoundError.Error()):
			return nil, entity.UsersNotFoundError
		case strings.Contains(err.Error(), entity.SearchingError.Error()):
			return nil, entity.SearchingError
		}
		return nil, err
	}

	followers := ReturnUsersList(followersList.Users)
	return followers, nil
}

func (followApp *FollowApp) GetAllFollowed(followerID int) ([]entity.User, error) {
	_, err := followApp.grpcClient.GetUser(context.Background(), &grpcUser.UserID{Uid: int64(followerID)})
	if err != nil {
		if strings.Contains(err.Error(), entity.UserNotFoundError.Error()) {
			return nil, entity.UserNotFoundError
		}
		return nil, err
	}

	followedList, err := followApp.grpcClient.GetAllFollowed(context.Background(), &grpcUser.UserID{Uid: int64(followerID)})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), entity.UsersNotFoundError.Error()):
			return nil, entity.UsersNotFoundError
		case strings.Contains(err.Error(), entity.SearchingError.Error()):
			return nil, entity.SearchingError
		}
		return nil, err
	}

	followed := ReturnUsersList(followedList.Users)
	return followed, nil
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
