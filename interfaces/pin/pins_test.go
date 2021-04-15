package pin

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
	"pinterest/interfaces/board"
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
	handleFunc   func(w http.ResponseWriter, r *http.Request)
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
var testBoardInfo board.BoardInfo

//headers := make(map[string][]string, 0)
//headers["Content-Length"] = []string{"zdes_dlina_tela"}
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
		"Testing first profile creation",
	},
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			nil,
			[]byte(`-----------------------------9051914041544843365972754266`+"\n" +
				`Content-Disposition: form-data;` +
				`name="pinInfo"` +"\n"+
				`{"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}`+"\n" +
				`-----------------------------9051914041544843365972754266`+"\n" +
				`Content-Disposition: form-data;` +"\n"+
				`name="pinImage"; ` +"\n"+
				`filename="a.txt"` +"\n"+
				`Content-Type: image/jpeg`+"\n" +
				`randomStr` +"\n"+
				`-----------------------------9051914041544843365972754266--`+"\n"),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			nil,
		},
		"Testing add first pin",
	},
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			nil,
			[]byte(`-----------------------------9051914041544843365972754266`+"\n" +
				`Content-Disposition: form-data;` +
				`name="pinInfo"` +"\n"+
				`{"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}`+"\n" +
				`-----------------------------9051914041544843365972754266`+"\n" +
				`Content-Disposition: form-data;` +"\n"+
				`name="pinImage"; ` +"\n"+
				`filename="a.txt"` +"\n"+
				`Content-Type: image/jpeg`+"\n" +
				`randomStr` +"\n"+
				`-----------------------------9051914041544843365972754266--`+"\n"),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			nil,
		},
		"Testing add second pin",
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
			[]byte(`{"title": "exampletitle1", "description": "exampleDescription1"}`),
		},
		"Testing add first board",
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
			[]byte(`{"ID":1,` +
				`"userID":0,` +
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
			[]byte(`[{"ID":0,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"},` +
				`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"pinImage":"example/link",` +
				`"description":"exampleDescription"}]`,
			),
		},
		"Testing get pin by board id",
	},
	{
		InputStruct{
			"/pin/add/1",
			"/pin/add/{id:[0-9]+}",
			"POST",
			nil,
			nil,
			testPinInfo.HandleSavePin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(nil),
		},
		"Testing saving second pin",
	},
	{
		InputStruct{
			"/board/0/add/0",
			"/board/{id:[0-9]+}/add/{pinID:[0-9]+}",
			"POST",
			nil,
			nil,
			testPinInfo.HandleAddPinToBoard,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(nil),
		},
		"Testing saving second pin to board",
	},
	{
		InputStruct{
			"/board/0/0",
			"/board/{id:[0-9]+}/{pinID:[0-9]+}",
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
			"/pin/3",
			"/pin/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			testPinInfo.HandleGetPinByID,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing get not existent pin by id",
	},
	{
		InputStruct{
			"/board/0/0",
			"/board/{id:[0-9]+}/{pinID:[0-9]+}",
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

var successCookies []*http.Cookie

func TestPins(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockS3App := mock_application.NewMockS3AppInterface(mockCtrl)
	mockUserApp := mock_application.NewMockUserAppInterface(mockCtrl)
	mockPinApp := mock_application.NewMockPinAppInterface(mockCtrl)
	//mockS3App := mock_application.NewMockS3AppInterface(mockCtrl)
	mockBoardApp := mock_application.NewMockBoardAppInterface(mockCtrl)
	cookieApp := application.NewCookieApp(40, 10*time.Hour)

	// TODO: maybe replace this with JSON parsing?
	expectedUserFirst := &entity.User{
		UserID:    0,
		Username:  "TestUsername",
		Password:  "thisisapassword",
		FirstName: "TestFirstName",
		LastName:  "TestLastName",
		Email:     "test@example.com",
		Avatar:    "avatars/1",
		Salt:      "",
	}

	mockUserApp.EXPECT().GetUserByUsername(expectedUserFirst.Username).Return(nil, fmt.Errorf("No user found with such username")).Times(1) // Handler will request user info
	mockUserApp.EXPECT().CreateUser(gomock.Any()).Return(expectedUserFirst.UserID, nil).Times(1)

	expectedPinFirst := &entity.Pin{
		PinId:       0,
		UserID:      0,
		Title:       "exampletitle",
		ImageLink:   "example/link",
		Description: "exampleDescription",
	}

	expectedBoardFirst := &entity.Board{
		BoardID:     0,
		UserID:      0,
		Title:       "exampletitle1",
		Description: "exampleDescription1",
	}

	expectedPinSecond := &entity.Pin{
		PinId:       1,
		UserID:      0,
		Title:       "exampletitle",
		ImageLink:   "example/link",
		Description: "exampleDescription",
	}

	expectedPinsInBoard := []entity.Pin{
		*expectedPinFirst,
		*expectedPinSecond,
	}
	mockBoardApp.EXPECT().CheckBoard(0, 0).Return(nil).Times(2)
	mockS3App.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockPinApp.EXPECT().CreatePin(expectedPinFirst).Return(expectedPinFirst.PinId, nil).Times(1)

	mockPinApp.EXPECT().CreatePin(gomock.Any()).Return(expectedPinSecond.PinId, nil).Times(1)

	mockBoardApp.EXPECT().AddBoard(expectedBoardFirst).Return(expectedBoardFirst.BoardID, nil).Times(1)

	mockPinApp.EXPECT().GetPin(expectedPinSecond.PinId).Return(expectedPinSecond, nil).Times(1)

	mockPinApp.EXPECT().GetPins(gomock.Any()).Return(expectedPinsInBoard, nil).Times(1)

	mockPinApp.EXPECT().SavePin(expectedUserFirst.UserID, expectedPinSecond.PinId).Return(nil).Times(1)

	mockBoardApp.EXPECT().CheckBoard(0, 0).Return(nil).Times(3)
	mockPinApp.EXPECT().AddPin(expectedBoardFirst.BoardID, expectedPinFirst.PinId).Return(nil).Times(1)

	mockPinApp.EXPECT().DeletePin(0, expectedPinFirst.PinId).Return(nil).Times(1)

	mockPinApp.EXPECT().GetPin(3).Return(nil, fmt.Errorf("No pin found")).Times(1)

	mockPinApp.EXPECT().DeletePin(expectedPinFirst.PinId, expectedUserFirst.UserID).Return(fmt.Errorf("pin not found")).Times(1)

	testAuthInfo = *auth.NewAuthInfo(mockUserApp, cookieApp, nil, nil) // We don't need S3 or board in these tests

	testBoardInfo = *board.NewBoardInfo(mockBoardApp)

	testPinInfo = PinInfo{
		pinApp:   mockPinApp,
		boardApp: mockBoardApp,
		s3App:    nil, // S3 is not needed, as we do not currently test file upload
	}
	for _, tt := range pinTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.handleFunc
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
