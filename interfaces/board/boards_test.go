package board

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
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
				`"title":"exampletitle",` +
				`"description":"exampleDescription"}`),
			testBoardInfo.HandleAddBoard,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"board_id": 0}`),
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
				`"title":"exampletitle",` +
				`"description":"exampleDescription"}`),
			testBoardInfo.HandleAddBoard,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"board_id": 1}`),
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
			[]byte(`{"boardID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"description":"exampleDescription"}`,
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
			[]byte(`[{"boardID":0,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"description":"exampleDescription"},` +
				`{"boardID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"description":"exampleDescription"}]`,
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

func TestProfileSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_application.NewMockUserAppInterface(mockCtrl)
	mockBoardApp := mock_application.NewMockBoardAppInterface(mockCtrl)

	cookieApp := application.NewCookieApp()

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

	mockUserApp.EXPECT().GetUserByUsername(gomock.Any()).Return(nil, fmt.Errorf("No user found with such username")).Times(1) // Handler will request user info
	mockUserApp.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedUser.UserID, nil).Times(1)

	expectedBoardFirst := entity.Board{
		BoardID:     0,
		UserID:      0,
		Title:       "exampletitle",
		Description: "exampleDescription",
	}

	expectedBoardSecond := entity.Board{
		BoardID:     1,
		UserID:      0,
		Title:       "exampletitle",
		Description: "exampleDescription",
	}

	expectedUserBoards := []entity.Board{
		expectedBoardFirst,
		expectedBoardSecond,
	}

	mockBoardApp.EXPECT().AddBoard(gomock.Any()).Return(expectedBoardFirst.BoardID, nil).Times(1)

	mockBoardApp.EXPECT().AddBoard(gomock.Any()).Return(expectedBoardSecond.BoardID, nil).Times(1)

	mockBoardApp.EXPECT().GetBoard(expectedBoardSecond.BoardID).Return(&expectedBoardSecond, nil).Times(1)

	mockBoardApp.EXPECT().GetBoards(expectedUser.UserID).Return(expectedUserBoards, nil).Times(1)

	mockBoardApp.EXPECT().DeleteBoard(expectedBoardFirst.BoardID, expectedUser.UserID).Return(nil).Times(1)

	mockBoardApp.EXPECT().GetBoard(3).Return(nil, fmt.Errorf("No board found")).Times(1)

	mockBoardApp.EXPECT().DeleteBoard(expectedBoardFirst.BoardID, expectedUser.UserID).Return(fmt.Errorf("pin not found")).Times(1)

	testAuthInfo = auth.AuthInfo{
		UserApp:      mockUserApp,
		CookieApp:    cookieApp,
		CookieLength: 40,
		Duration:     10 * time.Hour,
	}

	testBoardInfo = BoardInfo{
		BoardApp: mockBoardApp,
	}
	for _, tt := range boardTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.profileFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle, testAuthInfo.CookieApp)
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
