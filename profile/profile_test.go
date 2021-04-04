package profile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"pinterest/auth"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type profileInputStruct struct {
	url          string
	urlForRouter string
	method       string
	headers      map[string][]string
	postBody     []byte // JSON
	profileFunc  func(w http.ResponseWriter, r *http.Request)
	middleware   func(next http.HandlerFunc) http.HandlerFunc
}

// toHTTPRequest transforms profileInputStruct to http.Request, adding global cookies
func (input *profileInputStruct) toHTTPRequest(cookies []*http.Cookie) *http.Request {
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

type profileOutputStruct struct {
	responseCode int
	headers      map[string][]string
	postBody     []byte // JSON
}

// fillFromResponse transforms http.Response to profileOutputStruct
func (output *profileOutputStruct) fillFromResponse(response *http.Response) error {
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
var profileTestSuccess = []struct {
	in   profileInputStruct
	out  profileOutputStruct
	name string
}{
	{
		profileInputStruct{
			"/auth/signup",
			"/auth/signup",
			"POST",
			nil,
			[]byte(`{"username": "TestUsername",` +
				`"password": "thisisapassword",` +
				`"email": "test@example.com",` +
				`"firstName": "TestFirstName",` +
				`"lastName": "TestLastname",` +
				`"avatarLink": "avatars/1"}`,
			),
			auth.HandleCreateUser,
			auth.NoAuthMid,
		},

		profileOutputStruct{
			201,
			nil,
			nil,
		},
		"Testing profile creation",
	},
	{
		profileInputStruct{
			"/profile",
			"/profile",
			"GET",
			nil,
			nil,
			HandleGetProfile,
			auth.AuthMid, // If user is not logged in, they can't access their profile
		},

		profileOutputStruct{
			200,
			nil,
			[]byte(`{"username":"TestUsername",` + // No spaces because that's how go marshalls JSON
				`"email":"test@example.com",` +
				`"firstName":"TestFirstName",` +
				`"lastName":"TestLastname",` +
				`"avatarLink":"avatars/1"}`,
			),
		},
		"Testing profile output",
	},
	{
		profileInputStruct{
			"/profile/password",
			"/profile/password",
			"PUT",
			nil,
			[]byte(`{"password":"New Password"}`),
			HandleChangePassword,
			auth.AuthMid,
		},

		profileOutputStruct{
			200,
			nil,
			nil,
		},
		"Testing password change", // I don't know right now how to easily check if password changed
	},
	{
		profileInputStruct{
			"/profile/TestUsername",
			"/profile/{username}",
			"GET",
			nil,
			nil,
			HandleGetProfile,
			nil,
		},

		profileOutputStruct{
			200,
			nil,
			[]byte(`{"username":"TestUsername",` + // No spaces because that's how go marshalls JSON
				`"email":"test@example.com",` +
				`"firstName":"TestFirstName",` +
				`"lastName":"TestLastname",` +
				`"avatarLink":"avatars/1"}`,
			),
		},
		"Testing profile output using profile name",
	},
	{
		profileInputStruct{
			"/profile/edit",
			"/profile/edit",
			"PUT",
			nil,
			[]byte(`{"username": "new_User_Name",` +
				`"firstName": "new First name",` +
				`"lastName": "new Last Name",` +
				`"email": "new@example.com",` +
				`"avatarLink": "avatars/2"}`,
			),
			HandleEditProfile,
			auth.AuthMid,
		},

		profileOutputStruct{
			200,
			nil,
			nil,
		},
		"Testing profile edit",
	},
	{
		profileInputStruct{
			"/profile/0",
			"/profile/{id:[0-9]+}",
			"GET",
			nil,
			nil,
			HandleGetProfile,
			nil,
		},

		profileOutputStruct{
			200,
			nil,
			[]byte(`{"username":"new_User_Name",` +
				`"email":"new@example.com",` +
				`"firstName":"new First name",` +
				`"lastName":"new Last Name",` +
				`"avatarLink":"avatars/2"}`,
			),
		},
		"Testing profile output using profile id",
	},
	{
		profileInputStruct{
			"/profile/delete",
			"/profile/delete",
			"DELETE",
			nil,
			nil,
			HandleDeleteProfile,
			auth.AuthMid,
		},

		profileOutputStruct{
			200,
			nil,
			nil,
		},
		"Testing profile deletion",
	},
}

var successCookies []*http.Cookie

func TestProfileSuccess(t *testing.T) {
	for _, tt := range profileTestSuccess {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest(successCookies)

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			m := mux.NewRouter()
			funcToHandle := tt.in.profileFunc
			if tt.in.middleware != nil { // We don't always need middleware
				funcToHandle = tt.in.middleware(funcToHandle)
			}
			m.HandleFunc(tt.in.urlForRouter, funcToHandle).Methods(tt.in.method)
			m.ServeHTTP(rw, req)
			resp := rw.Result()

			// if server returned cookies, we use them
			if len(resp.Cookies()) > 0 {
				successCookies = resp.Cookies()
			}

			var result profileOutputStruct
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
