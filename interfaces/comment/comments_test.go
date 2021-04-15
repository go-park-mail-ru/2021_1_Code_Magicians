package comment

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

var testCommentInfo CommentInfo
var testAuthInfo auth.AuthInfo

var commentTest = []struct {
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
			"/comment/1",
			"/comment/{id:[0-9]+}",
			"POST",
			nil,
			[]byte(`{"pinID":1, "text":"Hello, my friends!!!"}`),
			testCommentInfo.HandleAddComment,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"text": "Hello, my friends!!!"}`),
		},
		"Testing add first comment",
	},
	{
		InputStruct{
			"/comment/1",
			"/comment/{id:[0-9]+}",
			"POST",
			nil,
			[]byte(`{"pinID":1,"text":"Welcome to the club, buddy!!!"}`),
			testCommentInfo.HandleAddComment,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"text": "Welcome to the club, buddy!!!"}`),
		},
		"Testing add second comment",
	},
	{
		InputStruct{
			"/comments/3",
			"/comments/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testCommentInfo.HandleGetComments,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing get not existent comments by pinID",
	},
	{
		InputStruct{
			"/comment/1",
			"/comment/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testCommentInfo.HandleGetComments,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"comments": [{"userID":0,"pinID":1,"text":"Hello, my friends!!!"},` +
				`{"userID":0,"pinID":1,"text":"Welcome to the club, buddy!!!"}]}`,
			),
		},
		"Testing get comments by pinID",
	},
	{
		InputStruct{
			"/comments/2",
			"/comments/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testCommentInfo.HandleGetComments,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"comments": []}`),
		},
		"Testing get not existent comments by pinID",
	},
}

var successCookies []*http.Cookie

func TestComments(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_application.NewMockUserAppInterface(mockCtrl)
	mockPinApp := mock_application.NewMockPinAppInterface(mockCtrl)
	mockCommentApp := mock_application.NewMockCommentAppInterface(mockCtrl)
	//mockS3App := mock_application.NewMockS3AppInterface(mockCtrl)

	cookieApp := application.NewCookieApp(40, 10*time.Hour)

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
	mockUserApp.EXPECT().CreateUser(gomock.Any()).Return(expectedUser.UserID, nil).Times(1)

	expectedPinFirst := entity.Pin{
		PinId:       1,
		Title:       "exampletitle",
		ImageLink:   "example/link",
		Description: "exampleDescription",
	}

	expectedPinSecond := entity.Pin{
		PinId:       2,
		Title:       "exampletitle",
		ImageLink:   "example/link",
		Description: "exampleDescription",
	}

	comment1 := entity.Comment{
		UserID:     0,
		PinID:      1,
		PinComment: "Hello, my friends!!!",
	}

	comment2 := entity.Comment{
		UserID:     0,
		PinID:      1,
		PinComment: "Welcome to the club, buddy!!!",
	}

	expectedComments := []entity.Comment{comment1, comment2}

	mockPinApp.EXPECT().GetPin(expectedPinFirst.PinId).Return(&expectedPinFirst, nil).Times(3)

	mockCommentApp.EXPECT().AddComment(gomock.Any()).Return(nil).Times(2)

	mockPinApp.EXPECT().GetPin(3).Return(nil, fmt.Errorf("No pin found")).Times(1)

	mockCommentApp.EXPECT().GetComments(expectedPinFirst.PinId).Return(expectedComments, nil)

	mockPinApp.EXPECT().GetPin(expectedPinSecond.PinId).Return(&expectedPinSecond, nil).Times(1)
	mockCommentApp.EXPECT().GetComments(expectedPinSecond.PinId).Return([]entity.Comment{}, nil)

	testAuthInfo = *auth.NewAuthInfo(mockUserApp, cookieApp, nil, nil) // We don't need S3 or board in these tests

	testCommentInfo = CommentInfo{
		pinApp:     mockPinApp,
		commentApp: mockCommentApp,
	}
	for _, tt := range commentTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.profileFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle, cookieApp)
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
