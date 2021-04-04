package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

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
var authTestSuccess = []struct {
	in   authInputStruct
	out  authOutputStruct
	name string
}{
	{
		authInputStruct{
			"/auth/signup",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"firstName": "TestFirstName",` +
				`"lastName": "TestLastname",` +
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
	},
	{
		authInputStruct{
			"/auth/logout",
			"GET",
			nil,
			nil,
			HandleLogoutUser,
			AuthMid,
		},

		authOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing user logout",
	},
	{
		authInputStruct{
			"/auth/login",
			"GET",
			nil,
			[]byte(`{"username": "TestUsername","password": "thisisapassword"}`),
			HandleLoginUser,
			NoAuthMid,
		},

		authOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing user login",
	},
	{
		authInputStruct{
			"/auth/check",
			"GET",
			nil,
			nil,
			HandleCheckUser,
			nil,
		},

		authOutputStruct{
			204,
			nil,
			nil,
		},
		"Testing checking if user is logged in when they are",
	},
}

var successCookies []*http.Cookie

func TestAuthSuccess(t *testing.T) {
	for _, tt := range authTestSuccess {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.authFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle)
			}
			m.HandleFunc(tt.in.url, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				successCookies = resp.Cookies()
			}

			var result authOutputStruct
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

// These tests have to run in that order!!!
var authTestFailure = []struct {
	in   authInputStruct
	out  authOutputStruct
	name string
}{
	{
		authInputStruct{
			"/auth/create",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername,` +
				`first_name": "TestFirstName",` +
				`"last_name": TestLastname",` +
				`"email": "test@example.com",` +
				`"password": "thisisapassword"`,
			),
			HandleCreateUser,
			NoAuthMid,
		},

		authOutputStruct{
			400,
			nil,
			nil,
		},
		"Testing wrong JSON when creating user",
	},
	{
		authInputStruct{
			"/auth/login",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername, password": "thisisapassword}}}`),
			HandleLoginUser,
			NoAuthMid,
		},

		authOutputStruct{
			400,
			nil,
			nil,
		},
		"Testing wrong JSON when logging user in",
	},
	{
		authInputStruct{
			"/auth/logout",
			"POST",
			nil,
			nil,
			HandleLogoutUser,
			AuthMid,
		},

		authOutputStruct{
			401,
			nil,
			nil,
		},
		"Testing trying to log user out without any cookies",
	},
	{
		authInputStruct{
			"/auth/check",
			"GET",
			nil,
			nil,
			HandleCheckUser,
			nil,
		},

		authOutputStruct{
			401,
			nil,
			nil,
		},
		"Testing checking if user is logged in when they aren't",
	},
}

var failureCookies []*http.Cookie

func TestAuthFailure(t *testing.T) {
	for _, tt := range authTestFailure {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(failureCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.authFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle)
			}
			m.HandleFunc(tt.in.url, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				failureCookies = resp.Cookies()
			}

			var result authOutputStruct
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
