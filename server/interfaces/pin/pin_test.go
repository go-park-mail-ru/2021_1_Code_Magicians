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
	handleFunc   func(w http.ResponseWriter, r *http.Request)
	middleware   func(next http.HandlerFunc, authApp application.AuthAppInterface) http.HandlerFunc
}

// toHTTPRequest transforms InputStruct to http.Request, adding global cookies
func (input *InputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
	reqURL, _ := url.Parse("https://localhost:8080" + input.url) // Scheme (https://) is required for URL parsing
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

var pinTest = []struct {
	in   InputStruct
	out  OutputStruct
	name string
}{
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			map[string][]string{
				"Content-Type": {"multipart/form-data; boundary=---------------------------9051914041544843365972754266"},
			},
			[]byte(`-----------------------------9051914041544843365972754266` + "\n" +
				`Content-Disposition: form-data; name="pinInfo"` + "\n" +
				"\n" +
				`{"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"description":"exampleDescription"}` + "\n" +
				`-----------------------------9051914041544843365972754266` + "\n" +
				`Content-Disposition: form-data; name="pinImage"; filename="a.jpg"` + "\n" +
				`Content-Type: image/jpeg` + "\n" +
				"\n" +
				`some image that is 1 black pixel` + "\n" +
				"\n" +
				`-----------------------------9051914041544843365972754266--` + "\n"),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, he can't post pins
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"ID":0}`),
		},
		"Testing add first pin",
	},
	{
		InputStruct{
			"/pin",
			"/pin",
			"POST",
			map[string][]string{
				"Content-Type": {"multipart/form-data; boundary=---------------------------9051914041544843365972754266"},
			},
			[]byte(`-----------------------------9051914041544843365972754266` + "\n" +
				`Content-Disposition: form-data; name="pinInfo"` + "\n" +
				"\n" +
				`{"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"description":"exampleDescription"}` + "\n" +
				`-----------------------------9051914041544843365972754266` + "\n" +
				`Content-Disposition: form-data; name="pinImage"; filename="a.jpg"` + "\n" +
				`Content-Type: image/jpeg` + "\n" +
				"\n" +
				`some image that is 1 black pixel` + "\n" +
				"\n" +
				`-----------------------------9051914041544843365972754266--` + "\n"),
			testPinInfo.HandleAddPin,
			middleware.AuthMid, // If user is not logged in, they can't access their profile
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"ID":1}`),
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
			[]byte(`{"ID":0}`),
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
				`"boardID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01T00:00:00Z",` +
				`"reportsCount":0}`,
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
			[]byte(`{"pins":[{"ID":0,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0},` +
				`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0}]}`,
			),
		},
		"Testing get pin by board id",
	},
	{
		InputStruct{
			"/pins/search?searchKey=exp&date=week",
			"/pins/search",
			"GET",
			nil,
			nil,
			testPinInfo.HandleSearchPins,
			middleware.AuthMid,
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"pins":[{"ID":0,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0},` +
				`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0}]}`,
			),
		},
		"Testing get pins by keyWords", // I don't know right now how to easily check if password changed
	},
	{
		InputStruct{
			"/pins/feed?offset=0&amount=10",
			"/pins/feed",
			"GET",
			nil,
			nil,
			testPinInfo.HandlePinsFeed,
			middleware.AuthMid,
		},

		OutputStruct{
			200,
			nil,
			[]byte(`{"pins":[{"ID":0,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0},` +
				`{"ID":1,` +
				`"userID":0,` +
				`"title":"exampletitle",` +
				`"imageLink":"example/link.jpg",` +
				`"imageHeight":1,` +
				`"imageWidth":1,` +
				`"imageAvgColor":"FFFFFF",` +
				`"description":"exampleDescription",` +
				`"creationDate":"0001-01-01 00:00:00 +0000 UTC",` +
				`"reportsCount":0}]}`,
			),
		},
		"Testing get pins for feed", // I don't know right now how to easily check if password changed
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
		"Testing remove pin from board", // I don't know right now how to easily check if password changed
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
		"Testing remove not existent pin from board",
	},
	{
		InputStruct{
			"/pin/report",
			"/pin/report",
			"POST",
			nil,
			[]byte(`{"pinID":0,` +
				`"description":"test report"}`,
			),
			testPinInfo.HandleCreateReport,
			middleware.AuthMid,
		},

		OutputStruct{
			201,
			nil,
			[]byte(`{"reportID":0}`),
		},
		"Testing adding report",
	},
	{
		InputStruct{
			"/pin/report",
			"/pin/report",
			"POST",
			nil,
			[]byte(`{"pinID":1234,` +
				`"description":"test report"}`,
			),
			testPinInfo.HandleCreateReport,
			middleware.AuthMid,
		},

		OutputStruct{
			404,
			nil,
			nil,
		},
		"Testing adding report for nonexistant pin",
	},
	{
		InputStruct{
			"/pin/report",
			"/pin/report",
			"POST",
			nil,
			[]byte(`{"pinID":0,` +
				`"description":"test report"}`,
			),
			testPinInfo.HandleCreateReport,
			middleware.AuthMid,
		},

		OutputStruct{
			409,
			nil,
			nil,
		},
		"Testing adding second report for same pin",
	},
}

var successCookies []*http.Cookie

func TestPins(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUserApp := mock_application.NewMockUserAppInterface(mockCtrl)
	mockAuthApp := mock_application.NewMockAuthAppInterface(mockCtrl)
	mockCookieApp := mock_application.NewMockCookieAppInterface(mockCtrl)
	mockPinApp := mock_application.NewMockPinAppInterface(mockCtrl)
	mockWebsocketApp := mock_application.NewMockWebsocketAppInterface(mockCtrl)
	mockBoardApp := mock_application.NewMockBoardAppInterface(mockCtrl)

	expectedUser := &entity.User{
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

	successCookies = nil
	successCookies = append(successCookies, &expectedCookie)

	mockAuthApp.EXPECT().CheckCookie(gomock.Any()).Return(&expectedCookieInfo, true).AnyTimes() // User is never logged out during these tests

	testLogger := zaptest.NewLogger(t)

	expectedPinFirst := &entity.Pin{
		PinID:         0,
		UserID:        0,
		Title:         "exampletitle",
		ImageLink:     "example/link.jpg",
		ImageHeight:   1,
		ImageWidth:    1,
		ImageAvgColor: "FFFFFF",
		Description:   "exampleDescription",
	}

	expectedBoardFirst := &entity.Board{
		BoardID:     0,
		UserID:      0,
		Title:       "exampletitle1",
		Description: "exampleDescription1",
	}

	expectedPinSecond := &entity.Pin{
		PinID:         1,
		UserID:        0,
		Title:         "exampletitle",
		ImageLink:     "example/link.jpg",
		ImageHeight:   1,
		ImageWidth:    1,
		ImageAvgColor: "FFFFFF",
		Description:   "exampleDescription",
	}

	expectedPinsInBoard := []entity.Pin{
		*expectedPinFirst,
		*expectedPinSecond,
	}

	mockPinApp.EXPECT().CreatePin(gomock.Any(), gomock.Any(), ".jpg").Return(expectedPinFirst.PinID, nil).Times(1)

	mockPinApp.EXPECT().CreatePin(gomock.Any(), gomock.Any(), ".jpg").Return(expectedPinSecond.PinID, nil).Times(1)

	mockBoardApp.EXPECT().CreateBoard(expectedBoardFirst).Return(expectedBoardFirst.BoardID, nil).Times(1)

	mockPinApp.EXPECT().GetPin(expectedPinSecond.PinID).Return(expectedPinSecond, nil).Times(1)

	mockPinApp.EXPECT().GetPins(gomock.Any()).Return(expectedPinsInBoard, nil).Times(1)

	mockPinApp.EXPECT().SearchPins("exp", "week").Return(expectedPinsInBoard, nil).Times(1)

	mockPinApp.EXPECT().GetPinsWithOffset(0, 10).Return(expectedPinsInBoard, nil).Times(1)

	mockPinApp.EXPECT().SavePin(expectedUser.UserID, expectedPinSecond.PinID).Return(nil).Times(1)

	mockBoardApp.EXPECT().CheckBoard(0, 0).Return(nil).Times(3)
	mockPinApp.EXPECT().AddPin(expectedBoardFirst.BoardID, expectedPinFirst.PinID).Return(nil).Times(1)

	mockPinApp.EXPECT().RemovePin(expectedBoardFirst.BoardID, expectedPinFirst.PinID).Return(nil).Times(1)

	mockPinApp.EXPECT().GetPin(3).Return(nil, entity.PinNotFoundError).Times(1)

	mockPinApp.EXPECT().RemovePin(expectedBoardFirst.BoardID, expectedPinFirst.PinID).Return(entity.PinNotFoundError).Times(1)

	mockPinApp.EXPECT().CreateReport(gomock.Any()).Return(0, nil).Times(1)

	mockPinApp.EXPECT().CreateReport(gomock.Any()).Return(-1, entity.PinNotFoundError).Times(1)

	mockPinApp.EXPECT().CreateReport(gomock.Any()).Return(-1, entity.DuplicateReportError).Times(1)

	testAuthInfo = *auth.NewAuthInfo(
		mockUserApp,
		mockAuthApp,
		mockCookieApp,
		nil, // We don't need S3 or board in these tests
		nil,
		mockWebsocketApp,
		testLogger)

	testBoardInfo = *board.NewBoardInfo(mockBoardApp, testLogger)

	testPinInfo = PinInfo{
		pinApp:   mockPinApp,
		boardApp: mockBoardApp,
		s3App:    nil, // S3 is not needed, as we do not currently test file upload
		logger:   testLogger,
	}
	for _, tt := range pinTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.handleFunc
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
