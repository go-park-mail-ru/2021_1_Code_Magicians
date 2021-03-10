package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/stretchr/testify/require"
)

type authInputStruct struct {
	url      string
	method   string
	headers  map[string][]string
	postBody []byte
	authFunc func(w http.ResponseWriter, r *http.Request)
}

// toHTTPRequest transforms authInputStruct to http.Request, adding global cookies
func (input *authInputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
	reqURL, _ := url.Parse(input.url)
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
			"localhost:8080/auth/create",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"first_name": "TestFirstName",` +
				`"last_name": "TestLastname",` +
				`"email": "test@example.com",` +
				`"password": "thisisapassword"}`,
			),
			HandleCreateUser,
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
			"localhost:8080/auth/login",
			"GET",
			nil,
			[]byte(`{"username": "TestUsername","password": "thisisapassword"}`),
			HandleLoginUser,
		},

		authOutputStruct{
			200,
			nil,
			nil,
		},
		"Testing user login",
	},
	{
		authInputStruct{
			"localhost:8080/auth/logout",
			"GET",
			nil,
			nil,
			HandleLogoutUser,
		},

		authOutputStruct{
			200,
			nil,
			nil,
		},
		"Testing user logout",
	},
}

var successCookies []*http.Cookie

func TestAuthSuccess(t *testing.T) {
	for _, tt := range authTestSuccess {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			tt.in.authFunc(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			successCookies = resp.Cookies()

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
			"localhost:8080/auth/create",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername,` +
				`first_name": "TestFirstName",` +
				`"last_name": TestLastname",` +
				`"email": "test@example.com",` +
				`"password": "thisisapassword"`,
			),
			HandleCreateUser,
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
			"localhost:8080/auth/login",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername, password": "thisisapassword}}}`),
			HandleLoginUser,
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
			"localhost:8080/auth/logout",
			"POST",
			nil,
			nil,
			HandleLoginUser,
		},

		authOutputStruct{
			400,
			nil,
			nil,
		},
		"Testing trying to log user out without any cookies",
	},
}

var failureCookies []*http.Cookie

func TestAuthFailure(t *testing.T) {
	for _, tt := range authTestFailure {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(failureCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			tt.in.authFunc(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			failureCookies = resp.Cookies()

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
