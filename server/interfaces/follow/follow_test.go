package follow

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"pinterest/domain/entity"
	"pinterest/interfaces/auth"
	"pinterest/usecase"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"pinterest/delivery/middleware"
	"pinterest/usecase/mock_usecase"
)

// followInputStruct stores information which will be parsed into request
type followInputStruct struct {
	url          string
	urlForRouter string
	method       string
	headers      map[string][]string
	postBody     []byte // JSON
	followFunc   func(w http.ResponseWriter, r *http.Request)
	middleware   func(next http.HandlerFunc, authApp usecase.AuthAppInterface) http.HandlerFunc
}

// toHTTPRequest transforms followInputStruct to http.Request, adding global cookies
func (input *followInputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
	reqURL, _ := url.Parse("http://localhost:8080" + input.url) // Scheme (http://) is required for URL parsing
	reqBody := bytes.NewBuffer(input.postBody)
	request := &http.Request{
		Method:        input.method,
		URL:           reqURL,
		Header:        input.headers,
		ContentLength: int64(reqBody.Len()),
		Body:          ioutil.NopCloser(reqBody),
	}

	if (len(cookies) > 0) && (request.Header == nil) {
		request.Header = make(http.Header)
	}

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	return request
}

// followOutputStruct stores information parsed from response
type followOutputStruct struct {
	responseCode int
	headers      map[string][]string
	postBody     []byte // JSON
}

// fillFromResponse transforms http.Response to followOutputStruct
func (output *followOutputStruct) fillFromResponse(response *http.Response) error {
	output.responseCode = response.StatusCode
	output.headers = response.Header
	if len(output.headers) == 0 {
		output.headers = nil
	}
	var err error
	output.postBody, err = ioutil.ReadAll(response.Body)
	if len(output.postBody) == 0 {
		output.postBody = nil
	}
	return err
}

var testFollowInfo FollowInfo
var testAuthInfo auth.AuthInfo

// These tests have to run in that order!!!
var followTestSuccess = []struct {
	in   followInputStruct
	out  followOutputStruct
	name string
}{
	{
		followInputStruct{
			"/auth/signup",
			"/auth/signup",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"password": "thisisapassword",` +
				`"email": "test@example.com",` +
				`"firstName": "TestFirstName",` +
				`"lastName": "TestLastName",` +
				`"avatarLink": "avatars/1"}`,
			),
			testAuthInfo.HandleCreateUser,
			middleware.NoAuthMid,
		},

		followOutputStruct{
			201,
			nil,
			nil,
		},
		"Testing profile creation",
	},
	{
		followInputStruct{
			"/follow/1",
			"/follow/{id:[0-9]+}",
			"POST",
			nil,
			nil,
			testFollowInfo.HandleFollowProfile,
			middleware.AuthMid,
		},

		followOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing following other profile using profile id",
	},
	{
		followInputStruct{
			"/follow/1",
			"/follow/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleUnfollowProfile,
			middleware.AuthMid,
		},

		followOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing unfollowing other profile using profile id",
	},
	{
		followInputStruct{
			"/follow/OtherUsername",
			"/follow/{username}",
			"POST",
			nil,
			nil,
			testFollowInfo.HandleFollowProfile,
			middleware.AuthMid,
		},

		followOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing following other profile using profile username",
	},
	{
		followInputStruct{
			"/follow/OtherUsername",
			"/follow/{username}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleUnfollowProfile,
			middleware.AuthMid,
		},

		followOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing unfollowing other profile using profile username",
	},
	{
		followInputStruct{
			"/following/0",
			"/following/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleGetFollowed,
			nil,
		},

		followOutputStruct{
			200,
			nil,
			[]byte(`{"profiles":[{"ID":0,` +
				`"username":"TestUsername",` +
				`"email":"test@example.com",` +
				`"firstName":"TestFirstName",` +
				`"lastName":"TestLastName",` +
				`"avatarLink":"avatars/1",` +
				`"following":0,` + // Follow counters are inconsistent, but it's no big deal
				`"followers":0}]}`,
			),
		},
		"Testing getting list of followed profiles",
	},
	{
		followInputStruct{
			"/followers/0",
			"/followers/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleGetFollowers,
			nil,
		},

		followOutputStruct{
			200,
			nil,
			[]byte(`{"profiles":[{"ID":0,` +
				`"username":"TestUsername",` +
				`"email":"test@example.com",` +
				`"firstName":"TestFirstName",` +
				`"lastName":"TestLastName",` +
				`"avatarLink":"avatars/1",` +
				`"following":0,` + // Follow counters are inconsistent, but it's no big deal
				`"followers":0}]}`,
			),
		},
		"Testing getting list of followers",
	},
}

var successCookies []*http.Cookie

func TestFollowSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_usecase.NewMockUserAppInterface(mockCtrl)
	mockAuthApp := mock_usecase.NewMockAuthAppInterface(mockCtrl)
	mockFollowApp := mock_usecase.NewMockFollowAppInterface(mockCtrl)
	mockNotificationApp := mock_usecase.NewMockNotificationAppInterface(mockCtrl)
	mockWebsocketApp := mock_usecase.NewMockWebsocketAppInterface(mockCtrl)
	testLogger := zaptest.NewLogger(t)

	// TODO: maybe replace this with JSON parsing?
	expectedUser := entity.User{
		UserID:    0,
		Username:  "TestUsername",
		Password:  "thisisapassword",
		FirstName: "TestFirstName",
		LastName:  "TestLastName",
		Email:     "test@example.com",
		Avatar:    "avatars/1",
		Salt:      "",
	}
	expectedCookie := http.Cookie{
		Name:     entity.CookieNameKey,
		Value:    "someRandomSessionValue",
		Path:     "/", // Cookie should be usable on entire website
		Expires:  time.Now().Add(10 * time.Hour),
		HttpOnly: true,
	}
	expectedCookieInfo := entity.CookieInfo{
		UserID: expectedUser.UserID,
		Cookie: &expectedCookie,
	}

	mockUserApp.EXPECT().CreateUser(gomock.Any()).Return(expectedUser.UserID, nil).Times(1)
	mockAuthApp.EXPECT().LoginUser(expectedUser.Username, expectedUser.Password).Return(&expectedCookieInfo, nil).Times(1)
	mockWebsocketApp.EXPECT().ChangeToken(expectedUser.UserID, "").Times(1)

	mockAuthApp.EXPECT().CheckCookie(gomock.Any()).Return(&expectedCookieInfo, true).AnyTimes() // User is never logged out during these tests, except for the last one

	expectedSecondUser := entity.User{
		UserID:    1,
		Username:  "OtherUsername",
		Password:  "thisisapassword",
		FirstName: "Other first name",
		LastName:  "Other last name",
		Email:     "other@example.com",
		Avatar:    "avatars/someotherpath",
		Salt:      "",
	}

	expectedUsers := []entity.User{expectedUser}

	notificationID := 0

	mockUserApp.EXPECT().GetUser(expectedSecondUser.UserID).Return(&expectedSecondUser, nil).Times(1) // HandleFollowProfile checks if followed profile exists
	mockFollowApp.EXPECT().Follow(expectedUser.UserID, expectedSecondUser.UserID).Return(nil).Times(1)
	mockUserApp.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(1) // HandleFollowProfile requests current user's username
	mockNotificationApp.EXPECT().AddNotification(gomock.Any()).Return(notificationID, nil).Times(1)
	mockNotificationApp.EXPECT().SendNotification(expectedSecondUser.UserID, notificationID).Return(nil).Times(1)

	mockUserApp.EXPECT().GetUser(expectedSecondUser.UserID).Return(&expectedSecondUser, nil).Times(1) // HandleUnfollowProfile checks if followed profile exists
	mockFollowApp.EXPECT().Unfollow(expectedUser.UserID, expectedSecondUser.UserID).Return(nil).Times(1)
	mockUserApp.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(1) // HandleUnfollowProfile requests current user's username
	mockNotificationApp.EXPECT().AddNotification(gomock.Any()).Return(notificationID, nil).Times(1)
	mockNotificationApp.EXPECT().SendNotification(expectedSecondUser.UserID, notificationID).Return(nil).Times(1)

	mockUserApp.EXPECT().GetUserByUsername(expectedSecondUser.Username).Return(&expectedSecondUser, nil).Times(1) // HandleFollowProfile checks if followed profile exists
	mockFollowApp.EXPECT().Follow(expectedUser.UserID, expectedSecondUser.UserID).Return(nil).Times(1)
	mockUserApp.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(1) // HandleFollowProfile requests current user's username
	mockNotificationApp.EXPECT().AddNotification(gomock.Any()).Return(0, nil).Times(1)
	mockNotificationApp.EXPECT().SendNotification(expectedSecondUser.UserID, notificationID).Return(nil).Times(1)

	mockUserApp.EXPECT().GetUserByUsername(expectedSecondUser.Username).Return(&expectedSecondUser, nil).Times(1) // HandleUnfollowProfile checks if followed profile exists
	mockFollowApp.EXPECT().Unfollow(expectedUser.UserID, expectedSecondUser.UserID).Return(nil).Times(1)
	mockUserApp.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(1) // HandleUnfollowProfile requests current user's username
	mockNotificationApp.EXPECT().AddNotification(gomock.Any()).Return(notificationID, nil).Times(1)
	mockNotificationApp.EXPECT().SendNotification(expectedSecondUser.UserID, notificationID).Return(nil).Times(1)

	mockFollowApp.EXPECT().GetAllFollowed(expectedUser.UserID).Return(expectedUsers, nil)

	mockFollowApp.EXPECT().GetAllFollowers(expectedUser.UserID).Return(expectedUsers, nil)

	testAuthInfo = *auth.NewAuthInfo(
		mockUserApp,
		mockAuthApp,
		nil,
		nil,
		mockWebsocketApp,
		testLogger,
	)

	testFollowInfo = FollowInfo{
		userApp:         mockUserApp,
		followApp:       mockFollowApp,
		notificationApp: mockNotificationApp,
		logger:          testLogger,
	}
	for _, tt := range followTestSuccess {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.followFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle, mockAuthApp)
			}
			m.HandleFunc(tt.in.urlForRouter, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				successCookies = resp.Cookies()
			}

			var result followOutputStruct
			result.fillFromResponse(resp)

			require.Equal(t, tt.out.responseCode, result.responseCode,
				fmt.Sprintf("Expected: %d as response code\nbut got:  %d",
					tt.out.responseCode, result.responseCode))
			for key, val := range tt.out.headers {
				resultVal, ok := result.headers[key]
				require.True(t, !ok,
					fmt.Sprintf("Expected header %s is not found:\nExpected: %v\nbut got: %v", key, tt.out.headers, result.headers))
				require.Equal(t, val, resultVal,
					fmt.Sprintf("Expected value of header %s: %v is different from actual value: %v", key, val, resultVal))
			}
			require.Equal(t, tt.out.postBody, result.postBody,
				fmt.Sprintf("Expected: %v as response body\nbut got:  %v",
					string(tt.out.postBody), string(result.postBody)))
		})
	}
}

// These tests have to run in that order!!!
var followTestFailure = []struct {
	in   followInputStruct
	out  followOutputStruct
	name string
}{
	{
		followInputStruct{
			"/following/0",
			"/following/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleGetFollowed,
			nil,
		},

		followOutputStruct{
			404,
			nil,
			nil,
		},
		"Testing getting empty list of followed profiles",
	},
	{
		followInputStruct{
			"/followers/0",
			"/followers/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testFollowInfo.HandleGetFollowers,
			nil,
		},

		followOutputStruct{
			404,
			nil,
			nil,
		},
		"Testing getting empty list of followers",
	},
}

var failureCookies []*http.Cookie

func TestFollowFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_usecase.NewMockUserAppInterface(mockCtrl)
	mockAuthApp := mock_usecase.NewMockAuthAppInterface(mockCtrl)
	mockFollowApp := mock_usecase.NewMockFollowAppInterface(mockCtrl)
	mockNotificationApp := mock_usecase.NewMockNotificationAppInterface(mockCtrl)
	mockWebsocketApp := mock_usecase.NewMockWebsocketAppInterface(mockCtrl)
	testLogger := zaptest.NewLogger(t)

	// TODO: maybe replace this with JSON parsing?
	expectedUser := entity.User{
		UserID:    0,
		Username:  "TestUsername",
		Password:  "thisisapassword",
		FirstName: "TestFirstName",
		LastName:  "TestLastName",
		Email:     "test@example.com",
		Avatar:    "avatars/1",
		Salt:      "",
	}

	mockFollowApp.EXPECT().GetAllFollowed(expectedUser.UserID).Return(nil, entity.UsersNotFoundError)

	mockFollowApp.EXPECT().GetAllFollowers(expectedUser.UserID).Return(nil, entity.UsersNotFoundError)

	testAuthInfo = *auth.NewAuthInfo(
		mockUserApp,
		mockAuthApp,
		nil,
		nil,
		mockWebsocketApp,
		testLogger,
	)

	testFollowInfo = FollowInfo{
		userApp:         mockUserApp,
		followApp:       mockFollowApp,
		notificationApp: mockNotificationApp,
		logger:          testLogger,
	}
	for _, tt := range followTestFailure {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(failureCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.followFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle, mockAuthApp)
			}
			m.HandleFunc(tt.in.urlForRouter, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				failureCookies = resp.Cookies()
			}

			var result followOutputStruct
			result.fillFromResponse(resp)

			require.Equal(t, tt.out.responseCode, result.responseCode,
				fmt.Sprintf("Expected: %d as response code\nbut got:  %d",
					tt.out.responseCode, result.responseCode))
			for key, val := range tt.out.headers {
				resultVal, ok := result.headers[key]
				require.True(t, !ok,
					fmt.Sprintf("Expected header %s is not found:\nExpected: %v\nbut got: %v", key, tt.out.headers, result.headers))
				require.Equal(t, val, resultVal,
					fmt.Sprintf("Expected value of header %s: %v is different from actual value: %v", key, val, resultVal))
			}
			require.Equal(t, tt.out.postBody, result.postBody,
				fmt.Sprintf("Expected: %v as response body\nbut got:  %v",
					string(tt.out.postBody), string(result.postBody)))
		})
	}
}
