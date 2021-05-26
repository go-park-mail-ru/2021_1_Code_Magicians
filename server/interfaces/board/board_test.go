package board

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"pinterest/application"
	"pinterest/application/mock_application"
	"pinterest/domain/entity"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/middleware"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/golang/mock/gomock"
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
	middleware   func(next http.HandlerFunc, authApp application.AuthAppInterface) http.HandlerFunc
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

var testBoardInfo BoardInfo
var testAuthInfo auth.AuthInfo

var boardTest = []struct {
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
			"/board",
			"/board",
			"POST",
			nil,
			[]byte(`{"userID":0,` +
				`"title":"exampletitle1",` +
				`"description":"exampleDescription1"}`),
			testBoardInfo.HandleCreateBoard,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"ID":0}`),
		},
		"Testing add first board",
	},
	{
		InputStruct{
			"/board",
			"/board",
			"POST",
			nil,
			[]byte(`{"userID":0,` +
				`"title":"exampletitle2",` +
				`"description":"exampleDescription2"}`),
			testBoardInfo.HandleCreateBoard,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"ID":1}`),
		},
		"Testing add second board",
	},
	{
		InputStruct{
			"/board/1",
			"/board/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testBoardInfo.HandleGetBoardByID,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle2",` +
				`"description":"exampleDescription2",` +
				`"avatarLink":"",` +
				`"avatarHeight":0,` +
				`"avatarWidth":0,` +
				`"avatarAvgColor":""}`,
			),
		},
		"Testing get board by boardID",
	},
	{
		InputStruct{
			"/boards/0",
			"/boards/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testBoardInfo.HandleGetBoardsByUserID,
			middleware.AuthMid,
		},
		OutputStruct{
			200,
			nil,
			[]byte(`{"boards":[{"ID":0,` +
				`"userID":0,` +
				`"title":"exampletitle1",` +
				`"description":"exampleDescription1",` +
				`"avatarLink":"",` +
				`"avatarHeight":0,` +
				`"avatarWidth":0,` +
				`"avatarAvgColor":""},` +
				`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle2",` +
				`"description":"exampleDescription2",` +
				`"avatarLink":"",` +
				`"avatarHeight":0,` +
				`"avatarWidth":0,` +
				`"avatarAvgColor":""}]}`,
			),
		},
		"Testing get boards by user id",
	},
	{
		InputStruct{
			"/board/0",
			"/board/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testBoardInfo.HandleDelBoardByID,
			middleware.AuthMid,
		},

		OutputStruct{
			204,
			nil,
			nil,
		},
		"Testing delete board", // I don't know right now how to easily check if password changed
	},
	{
		InputStruct{
			"/board/3",
			"/board/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testBoardInfo.HandleGetBoardByID,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing get not existent board by boardID",
	},
	{
		InputStruct{
			"/board/0",
			"/board/{id:[0-9]+}",
			"DELETE",
			nil,
			nil,
			testBoardInfo.HandleDelBoardByID,
			middleware.AuthMid,
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing delete not existent board",
	},
}

var successCookies []*http.Cookie

func TestBoards(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_application.NewMockUserAppInterface(mockCtrl)
	mockAuthApp := mock_application.NewMockAuthAppInterface(mockCtrl)
	mockCookieApp := mock_application.NewMockCookieAppInterface(mockCtrl)
	mockBoardApp := mock_application.NewMockBoardAppInterface(mockCtrl)
	mockWebsocketApp := mock_application.NewMockWebsocketAppInterface(mockCtrl)
	testLogger := zaptest.NewLogger(t)

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

	mockCookieApp.EXPECT().GenerateCookie().Return(&expectedCookie, nil).Times(1)
	mockUserApp.EXPECT().CreateUser(gomock.Any()).Return(expectedUser.UserID, nil).Times(1)
	mockWebsocketApp.EXPECT().ChangeToken(expectedUser.UserID, "").Times(1)
	mockCookieApp.EXPECT().AddCookieInfo(gomock.Any()).Return(nil).Times(1)

	mockAuthApp.EXPECT().CheckCookie(gomock.Any()).Return(&expectedCookieInfo, true).AnyTimes() // User is never logged out during these tests

	expectedBoardFirst := entity.Board{
		BoardID:     0,
		UserID:      0,
		Title:       "exampletitle1",
		Description: "exampleDescription1",
	}

	expectedBoardSecond := entity.Board{
		BoardID:     1,
		UserID:      0,
		Title:       "exampletitle2",
		Description: "exampleDescription2",
	}

	boardInfo1 := entity.Board{
		BoardID:     1,
		UserID:      0,
		Title:       "exampletitle2",
		Description: "exampleDescription2",
	}
	expectedUserBoards := []entity.Board{
		expectedBoardFirst,
		expectedBoardSecond,
	}

	mockBoardApp.EXPECT().CreateBoard(gomock.Any()).Return(expectedBoardFirst.BoardID, nil).Times(1)

	mockBoardApp.EXPECT().CreateBoard(gomock.Any()).Return(expectedBoardSecond.BoardID, nil).Times(1)

	mockBoardApp.EXPECT().GetBoard(expectedBoardSecond.BoardID).Return(&boardInfo1, nil).Times(1)

	mockBoardApp.EXPECT().GetBoards(expectedUser.UserID).Return(expectedUserBoards, nil).Times(1)

	mockBoardApp.EXPECT().DeleteBoard(expectedUser.UserID, expectedBoardFirst.BoardID).Return(nil).Times(1)

	mockBoardApp.EXPECT().GetBoard(3).Return(nil, entity.BoardNotFoundError).Times(1)

	mockBoardApp.EXPECT().DeleteBoard(expectedUser.UserID, expectedBoardFirst.BoardID).Return(entity.BoardNotFoundError).Times(1)

	testAuthInfo = *auth.NewAuthInfo(
		mockUserApp,
		mockAuthApp,
		mockCookieApp,
		nil, // We don't need S3 in these tests
		mockBoardApp,
		mockWebsocketApp,
		testLogger,
	)

	testBoardInfo = BoardInfo{
		boardApp: mockBoardApp,
		logger:   testLogger,
	}
	for _, tt := range boardTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.profileFunc
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
					string(tt.out.postBody), string(result.postBody)))
		})
	}
}
