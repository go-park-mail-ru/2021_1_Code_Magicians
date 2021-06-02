package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"pinterest/domain/entity"
	grpcAuth "pinterest/services/auth/proto"
	_ "pinterest/services/user/proto"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthApp struct {
	grpcClient grpcAuth.AuthClient
	us         UserAppInterface
	cookieApp  CookieAppInterface

	vkClientID     string
	vkClientSecret string
}

func NewAuthApp(grpcClient grpcAuth.AuthClient, us UserAppInterface, cookieApp CookieAppInterface,
	vkClientID string, vkClientSecret string) *AuthApp {
	return &AuthApp{
		grpcClient:     grpcClient,
		us:             us,
		cookieApp:      cookieApp,
		vkClientID:     vkClientID,
		vkClientSecret: vkClientSecret,
	}
}

type AuthAppInterface interface {
	CheckUserCredentials(username string, password string) (*entity.CookieInfo, error)
	LogoutUser(userID int) error
	CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool)                      // Check if passed cookie value is present in any active session
	CheckVkCode(code string, redirectURI string) (*entity.CookieInfo, error)         // Use public vk token to get private one and log user with that token in
	AddVkCode(userID int, code string, redirectURI string) error                     // Use public vk token to get private one and associate it with user
	VkCodeToToken(code string, redirectURI string) (*entity.UserVkTokenInput, error) // Get private token from vk using code
}

func (authApp *AuthApp) CheckUserCredentials(username string, password string) (*entity.CookieInfo, error) {
	user, err := authApp.us.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// TODO: rework, check password emptyness

	_, err = authApp.grpcClient.CheckUserCredentials(context.Background(),
		&grpcAuth.UserAuth{Username: username, Password: password})

	if err != nil { // TODO: hashing
		if strings.Contains(err.Error(), entity.IncorrectPasswordError.Error()) {
			return nil, entity.IncorrectPasswordError
		}
		return nil, err
	}

	cookie, err := authApp.cookieApp.GenerateCookie()
	if err != nil {
		for err == entity.DuplicatingCookieValueError {
			cookie, err = authApp.cookieApp.GenerateCookie()
		}
	}

	resultCookieInfo := &entity.CookieInfo{UserID: user.UserID, Cookie: cookie}
	err = authApp.cookieApp.AddCookieInfo(resultCookieInfo)
	if err != nil {
		return nil, err
	}

	return resultCookieInfo, nil
}

func (authApp *AuthApp) LogoutUser(userID int) error {
	cookieInfo, found := authApp.cookieApp.SearchByUserID(userID)
	if !found {
		return entity.UserNotLoggedInError
	}

	return authApp.cookieApp.RemoveCookie(cookieInfo)
}

func (authApp *AuthApp) CheckCookie(cookie *http.Cookie) (*entity.CookieInfo, bool) {
	grpcCookieInfo, isCookie := authApp.cookieApp.SearchByValue(cookie.Value)

	if !isCookie {
		return nil, false
	}

	return &entity.CookieInfo{
		UserID: int(grpcCookieInfo.UserID),
		Cookie: cookie,
	}, isCookie
}

func (authApp *AuthApp) CheckVkCode(code string, redirectURI string) (*entity.CookieInfo, error) {
	tokenInput, err := authApp.VkCodeToToken(code, redirectURI)
	if err != nil {
		return nil, err
	}

	grpcUserID, err := authApp.grpcClient.CheckUserByVkToken(context.Background(),
		&grpcAuth.VkToken{Token: tokenInput.Token})

	//TODO: fix behaviour when token expires and is replaced in vk's response

	if err != nil {
		if strings.Contains(err.Error(), entity.VkTokenNotFoundError.Error()) {
			return nil, entity.VkTokenNotFoundError
		}
		return nil, err
	}

	cookie, err := authApp.cookieApp.GenerateCookie()
	for err == entity.DuplicatingCookieValueError {
		cookie, err = authApp.cookieApp.GenerateCookie()
	}

	resultCookieInfo := &entity.CookieInfo{UserID: int(grpcUserID.Uid), Cookie: cookie}
	err = authApp.cookieApp.AddCookieInfo(resultCookieInfo)
	if err != nil {
		return nil, err
	}

	return resultCookieInfo, nil
}

func (authApp *AuthApp) AddVkCode(userID int, code string, redirectURI string) error {
	tokenInput, err := authApp.VkCodeToToken(code, redirectURI)
	if err != nil {
		return err
	}

	_, err = authApp.grpcClient.AddVkToken(context.Background(),
		&grpcAuth.VkTokenInfo{UserID: int64(userID), Token: tokenInput.Token, Expires: secondsLeftToTimestamp(tokenInput.Expires)})

	if err != nil {
		if strings.Contains(err.Error(), entity.VkTokenDuplicateError.Error()) {
			return entity.VkTokenDuplicateError
		}
		return err
	}

	return nil
}

func (authApp *AuthApp) VkCodeToToken(code string, redirectURI string) (*entity.UserVkTokenInput, error) {
	resp, err := http.Get(fmt.Sprintf("%saccess_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s",
		entity.VkAuthURLKey, authApp.vkClientID, authApp.vkClientSecret, redirectURI, code))
	if err != nil {
		// TODO: error handling
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(data))
		// TODO: proper vk errors handling
		return nil, fmt.Errorf("Could not get token from vk")
	}

	userTokenInput := new(entity.UserVkTokenInput)
	err = json.NewDecoder(resp.Body).Decode(userTokenInput)
	if err != nil {
		return nil, err
	}

	return userTokenInput, nil
}

// return timestamp seconds away from now
func secondsLeftToTimestamp(seconds int) *timestamppb.Timestamp {
	return timestamppb.New(time.Now().Add(time.Duration(int64(seconds) * 1000)))
}
