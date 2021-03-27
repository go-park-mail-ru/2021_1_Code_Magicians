package pins

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
 . "pinterest/auth"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type authInputStruct struct {
	url        string
	method     string
	headers    map[string][]string
	postBody   []byte
	authFunc   func(w http.ResponseWriter, r *http.Request)
	middleware func(next http.HandlerFunc) http.HandlerFunc
}

// toHTTPRequest transforms authInputStruct to http.Request, adding global cookies
func (input *authInputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
	reqURL, _ := url.Parse("https://localhost:8080" + input.url)
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

type authOutputStruct struct {
	responseCode int
	headers      map[string][]string
	postBody     []byte
}

// fillFromResponse transforms http.Response to authOutputStruct
func (output *authOutputStruct) fillFromResponse(response *http.Response) error {
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

// These tests have to run in that order!!!
type authTest struct {
	in   authInputStruct
	out  authOutputStruct
	name string
}


var successCookies []*http.Cookie

//func TestAuthSuccess(t *testing.T) {
//	for _, tt := range authTest {
//		tt := tt
//		t.Run(tt.name, func(t *testing.T) {
//			req := tt.in.toHTTPRequest(successCookies)
//
//			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
//			m := mux.NewRouter()
//			funcToHandle := tt.in.authFunc
//			if tt.in.middleware != nil { // We don't always need middleware
//				funcToHandle = tt.in.middleware(funcToHandle)
//			}
//			m.HandleFunc(tt.in.url, funcToHandle).Methods(tt.in.method)
//			m.ServeHTTP(rw, req)
//			resp := rw.Result()
//
//			// if server returned cookies, we use them
//			if len(resp.Cookies()) > 0 {
//				successCookies = resp.Cookies()
//			}
//
//			var result authOutputStruct
//			result.fillFromResponse(resp)
//
//			require.Equal(t, tt.out.responseCode, result.responseCode,
//				fmt.Sprintf("Expected: %d as response code\nbut got:  %d",
//					tt.out.responseCode, result.responseCode))
//			for key, val := range tt.out.headers {
//				resultVal, ok := result.headers[key]
//				require.True(t, !ok,
//					fmt.Sprintf("Expected header %s is not found:\nExpected: %v\nbut got: %v", key, tt.out.headers, result.headers))
//				require.Equal(t, val, resultVal,
//					fmt.Sprintf("Expected value of header %s: %v is different from actual value: %v", key, val, resultVal))
//			}
//			require.Equal(t, tt.out.postBody, result.postBody,
//				fmt.Sprintf("Expected: %v as response body\nbut got:  %v",
//					tt.out.postBody, result.postBody))
//		})
//	}
//}

var testPinSet = PinsStorage{
	Storage: NewPinsSet(),
}

func TestUserPinSet_AddPin(t *testing.T) {
	authTest := authTest{
		authInputStruct{
			"/auth/create",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"first_name": "TestFirstName",` +
				`"last_name": "TestLastname",` +
				`"email": "test@example.com",` +
				`"password": "thisisapassword"}`,
			),
			HandleCreateUser,
			NoAuthMid,
		},

			authOutputStruct{
				201,
				nil,
				nil,
			},
			"Testing user creation",
	}

	t.Run(authTest.name, func(t *testing.T) {
			req := authTest.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := authTest.in.authFunc
			if authTest.in.middleware != nil { // We don't always need middleware
				funcToHandle = authTest.in.middleware(funcToHandle)
			}
			m.HandleFunc(authTest.in.url, funcToHandle).Methods(authTest.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				successCookies = resp.Cookies()
			}

			var result authOutputStruct
			result.fillFromResponse(resp)

			require.Equal(t, authTest.out.responseCode, result.responseCode,
				fmt.Sprintf("Expected: %d as response code\nbut got:  %d",
					authTest.out.responseCode, result.responseCode))
			for key, val := range authTest.out.headers {
				resultVal, ok := result.headers[key]
				require.True(t, !ok,
					fmt.Sprintf("Expected header %s is not found:\nExpected: %v\nbut got: %v", key, authTest.out.headers, result.headers))
				require.Equal(t, val, resultVal,
					fmt.Sprintf("Expected value of header %s: %v is different from actual value: %v", key, val, resultVal))
			}
			require.Equal(t, authTest.out.postBody, result.postBody,
				fmt.Sprintf("Expected: %v as response body\nbut got:  %v",
					authTest.out.postBody, result.postBody))
			boardID := 0

			body := strings.NewReader(
				fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "pinImage": "example/link", "description": "exampleDescription"}`, boardID),
			)
			r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/", body)
			w := httptest.NewRecorder()
			testPinSet.Storage.AddPin(w, r)

			body = strings.NewReader(
				fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "pinImage": "example/link", "description": "exampleDescription"}`, boardID),
			)
			r = httptest.NewRequest("POST", "http://127.0.0.1:8080/pin", body)
			w = httptest.NewRecorder()
			testPinSet.Storage.AddPin(w, r)

			resResponse := w.Result()
			resBody, _ := io.ReadAll(resResponse.Body)

			require.Equal(t, http.StatusCreated, resResponse.StatusCode)
			//require.Equal(t, resResponse.Header.Get("Content-Type"), "text/plain; charset=utf-8")
			require.Equal(t, string(resBody), "{\"pin_id\": 1}")
			require.Equal(t, len(testPinSet.Storage.userPins[0]), 2)
		})


}

func TestUserPinSet_AddPinBadData(t *testing.T) {
	body := strings.NewReader(
		fmt.Sprintf(`{"bad data"}`),
	)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/", body)
	w := httptest.NewRecorder()
	testPinSet.Storage.AddPin(w, r)

	resResponse := w.Result()

	require.Equal(t, http.StatusBadRequest, resResponse.StatusCode)
}

func TestUserPinSet_GetPinByID(t *testing.T) {
	expectedResponse := `{"id":1,"boardID":0,"title":"exampletitle","pinImage":"example/link","description":"exampleDescription"}`
	boardID := 0
	body := strings.NewReader(
		fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "pinImage": "example/link", "description": "exampleDescription"}`, boardID),
	)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin", body)
	w := httptest.NewRecorder()

	testPinSet.Storage.AddPin(w, r)
	body = strings.NewReader(
		fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "pinImage": "example/link", "description": "exampleDescription"}`, boardID),
	)
	r = httptest.NewRequest("POST", "http://127.0.0.1:8080/pin", body)
	w = httptest.NewRecorder()
	testPinSet.Storage.AddPin(w, r)

	r = httptest.NewRequest("GET", "http://127.0.0.1:8080/pins/0", nil)
	w = httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "1"})

	testPinSet.Storage.GetPinByID(w, r)
	resResponse := w.Result()
	resBody, _ := io.ReadAll(resResponse.Body)
	require.Equal(t, http.StatusOK, resResponse.StatusCode)
	require.Equal(t, string(resBody), expectedResponse)
}

func TestUserPinSet_GetPinByIDError(t *testing.T) {
	r := httptest.NewRequest("GET", "http://127.0.0.1:8080/pins/4", nil)
	w := httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "4"})

	testPinSet.Storage.GetPinByID(w, r)
	resResponse := w.Result()

	require.Equal(t, http.StatusNotFound, resResponse.StatusCode)
}

func TestUserPinSet_DelPinByID(t *testing.T) {
	r := httptest.NewRequest("DELETE", "http://127.0.0.1:8080/pins/0", nil)
	w := httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "0"})
	testPinSet.Storage.DelPinByID(w, r)
	resResponse := w.Result()

	require.Equal(t, http.StatusOK, resResponse.StatusCode)
	require.Equal(t, len(testPinSet.Storage.userPins[0]), 3)
}

func TestUserPinSet_DelNoSuchPin(t *testing.T) {
	r := httptest.NewRequest("DELETE", "http://127.0.0.1:8080/pins/0", nil)
	w := httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "0"})

	testPinSet.Storage.DelPinByID(w, r)
	resResponse := w.Result()

	require.Equal(t, http.StatusNotFound, resResponse.StatusCode)
}

func TestUserPinSet_BadIdCase1(t *testing.T) {
	r := httptest.NewRequest("DELETE", "http://127.0.0.1:8080/pins/1", nil)
	w := httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "badId"})

	testPinSet.Storage.GetPinByID(w, r)
	resResponse := w.Result()

	require.Equal(t, http.StatusBadRequest, resResponse.StatusCode)
}

func TestUserPinSet_BadIdCase2(t *testing.T) {
	r := httptest.NewRequest("DELETE", "http://127.0.0.1:8080/pins/1", nil)
	w := httptest.NewRecorder()
	r = mux.SetURLVars(r, map[string]string{"id": "badId"})

	testPinSet.Storage.DelPinByID(w, r)
	resResponse := w.Result()

	require.Equal(t, http.StatusBadRequest, resResponse.StatusCode)
}
