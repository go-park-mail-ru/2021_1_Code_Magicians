package pin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"pinterest/application"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/middleware"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type InputStruct struct {
	url          string
	urlForRouter string
	method       string
	headers      map[string][]string
	postBody     []byte // JSON
	profileFunc  func(w http.ResponseWriter, r *http.Request)
	middleware   func(next http.HandlerFunc, cookieApp application.CookieAppInterface) http.HandlerFunc
}

// toHTTPRequest transforms InputStruct to http.Request, adding global cookies
func (input *InputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
	reqURL, _ := url.Parse("https://localhost:8080" + input.url) // Scheme (https://) is required for URL parsing
	reqBody := bytes.NewBuffer(input.postBody)
	request := &http.Request{
		Method: input.method,
		URL:    reqURL,
		Header: input.headers,
		Body:   ioutil.NopCloser(reqBody),
	}

	if (len(cookies) > 0) && (request.Header == nil) {
		request.Header = make(http.Header)
	}

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	return request
}

type OutputStruct struct {
	responseCode int
	headers      map[string][]string
	postBody     []byte // JSON
}

// fillFromResponse transforms http.Response to OutputStruct
func (output *OutputStruct) fillFromResponse(response *http.Response) error {
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

var testPinInfo PinInfo
var testAuthInfo auth.AuthInfo

var successCookies []*http.Cookie

var pinTest = []struct {
	in   InputStruct
	out  OutputStruct
	name string
}{
	{
		InputStruct{
			"/auth/signup",
			"/auth/signup",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"password": "thisisapassword",` +
				`"first_name": "TestFirstName",` +
				`"last_name": "TestLastname",` +
				`"email": "test@example.com",` +
				`"avatar": "avatars/1"}`,
			),
			testAuthInfo.HandleCreateUser,
			middleware.NoAuthMid,
		},

		OutputStruct{
			201,
			nil,
			nil,
		},
		"Testing profile creation",
	},
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			nil,
			[]byte(`{"boardID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}`),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"pin_id": 0}`),
		},
		"Testing add first pin",
	},
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			nil,
			[]byte(`{"boardID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}`),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"pin_id": 1}`),
		},
		"Testing add second pin",
	},
	{
		InputStruct{
			"/pin/1",
			"/pin/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testPinInfo.HandleGetPinByID,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"id":1,` +
				`"boardID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}`,
			),
		},
		"Testing get pin by id",
	},
	{
		InputStruct{
			"/pins/0",
			"/pins/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testPinInfo.HandleGetPinsByBoardID,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			200,
			nil,
			[]byte(`[{"id":0,` +
				`"boardID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"},` +
				`{"id":1,` +
				`"boardID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}]`,
			),
		},
		"Testing get pin by board id",
	},
	{
		InputStruct{
			"/pin/0",
			"/pin/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testPinInfo.HandleDelPinByID,
			middleware.AuthMid,
		},

		OutputStruct{
			204,
			nil,
			nil,
		},
		"Testing delete pin", // I don't know right now how to easily check if password changed
	},
	{
		InputStruct{
			"/pins/3",
			"/pin/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testPinInfo.HandleGetPinByID,
			middleware.NoAuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			404,
			nil,
			[]byte("404 page not found\n"),
		},
		"Testing get not existent pin by id",
	},
	{
		InputStruct{
			"/pins/0",
			"/pins/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testPinInfo.HandleDelPinByID,
			middleware.AuthMid,
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing delete not existent pin", // I don't know right now how to easily check if password changed
	},
}

func TestUserPins(t *testing.T) {
	//mockCtrl := gomock.NewController(t)
	//defer mockCtrl.Finish()
	//mockDoer := mock_repository.NewMockUserRepository(mockCtrl)
	////
	//expectedUser := entity.User{
	//	UserID:    0,
	//	Username:  "TestUsername",
	//	Password:  "thisisapassword",
	//	FirstName: "TestFirstName",
	//	LastName:  "TestLastName",
	//	Email:     "test@example.com",
	//	Avatar:    "avatars/1",
	//	Salt:      "",
	//}
	//mockDoer.EXPECT().GetUserByUsername(expectedUser.Username).Return(nil, nil).Times(1) // Credentials check
	//mockDoer.EXPECT().CreateUser(gomock.Any()).Return(expectedUser.UserID, nil).Times(1)
	//
	//mockDoer.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(2) // Credentials check, then normal user output using cookie's userID
	//
	//mockDoer.EXPECT().SaveUser(gomock.Any()).Return(nil).Times(1)
	//
	//expectedUser.Password = "New Password"
	//mockDoer.EXPECT().GetUserByUsername(expectedUser.Username).Return(&expectedUser, nil).Times(1) // Normal user output using username
	//
	//mockDoer.EXPECT().GetUser(expectedUser.UserID).Return(&expectedUser, nil).Times(1) // Credentials check
	//mockDoer.EXPECT().SaveUser(gomock.Any()).Return(nil).Times(1)
	//
	//expecteduser := entity.User{
	//	UserID:    0,
	//	Username:  "new_User_Name",
	//	Password:  "New Password",
	//	FirstName: "new First name",
	//	LastName:  "new Last Name",
	//	Email:     "new@example.com",
	//	Avatar:    "avatars/2",
	//	Salt:      "",
	//}
	//mockDoer.EXPECT().GetUser(expecteduser.UserID).Return(&expecteduser, nil).Times(1) // Normal user output using userID
	//
	//mockDoer.EXPECT().DeleteUser(expecteduser.UserID).Return(nil).Times(1)

	userApp := application.NewUserApp(nil)
	cookieApp := application.NewCookieApp()

	testAuthInfo = auth.AuthInfo{
		UserApp:      userApp,
		CookieApp:    cookieApp,
		CookieLength: 40,
		Duration:     10 * time.Hour,
	}
	counter := 0
	for _, tt := range pinTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.profileFunc
			if tt.in.middleware != nil { // We don't always need middleware
				if counter == 0 {
					funcToHandle = tt.in.middleware(funcToHandle, testAuthInfo.CookieApp)
					counter++
				} else {
					funcToHandle = tt.in.middleware(funcToHandle, nil)
				}
			}
			m.HandleFunc(tt.in.urlForRouter, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				successCookies = resp.Cookies()
			}

			var result OutputStruct
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
					tt.out.postBody, result.postBody))
		})
	}
}
