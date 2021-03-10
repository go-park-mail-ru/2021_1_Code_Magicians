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

func (input *authInputStruct) toHTTPRequest() *http.Request {
	reqURL, _ := url.Parse(input.url)
	reqBody := bytes.NewBuffer(input.postBody)
	request := &http.Request{
		Method: input.method,
		URL:    reqURL,
		Header: input.headers,
		Body:   ioutil.NopCloser(reqBody),
	}
	return request
}

type authOutputStruct struct {
	ResponseCode int
	Headers      map[string][]string
	PostBody     []byte
}

func (output *authOutputStruct) fillFromResponse(response *http.Response) error {
	output.ResponseCode = response.StatusCode
	output.Headers = response.Header
	if len(output.Headers) == 0 {
		output.Headers = nil
	}
	var err error
	output.PostBody, err = ioutil.ReadAll(response.Body)
	if len(output.PostBody) == 0 {
		output.PostBody = nil
	}
	return err
}

var cookie http.Cookie
var authTest = []struct {
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
}

func TestAuth(t *testing.T) {
	for _, tt := range authTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := tt.in.toHTTPRequest()
			// requestDump, err := httputil.DumpRequest(req, true)
			// if err != nil {
			// 	t.Log(err)
			// }
			// t.Log(string(requestDump))

			rw := httptest.NewRecorder() // not ResponseWriter because we need to read response
			tt.in.authFunc(rw, req)
			resp := rw.Result()

			// responseDump, err := httputil.DumpResponse(resp, true)
			// if err != nil {
			// 	t.Log(err)
			// }
			// t.Log(string(responseDump))
			var result authOutputStruct
			result.fillFromResponse(resp)

			require.Equal(t, tt.out.ResponseCode, result.ResponseCode,
				fmt.Sprintf("Expected: %d as response code\nbut got:  %d",
					tt.out.ResponseCode, result.ResponseCode))
			for key, val := range tt.out.Headers {
				resultVal, ok := result.Headers[key]
				require.True(t, !ok,
					fmt.Sprintf("Expected header %s is not found:\nExpected: %v\nbut got: %v", key, tt.out.Headers, result.Headers))
				require.Equal(t, val, resultVal,
					fmt.Sprintf("Expected value of header %s: %v is different from actual value: %v", key, val, resultVal))
			}
			require.Equal(t, tt.out.PostBody, result.PostBody,
				fmt.Sprintf("Expected: %v as response body\nbut got:  %v",
					tt.out.PostBody, result.PostBody))
		})
	}
}
