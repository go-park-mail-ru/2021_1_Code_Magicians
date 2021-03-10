package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/google/go-cmp/cmp"
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
	responseCode int
	headers      map[string][]string
	postBody     []byte
}

func (output *authOutputStruct) fillFromResponse(response *http.Response) error {
	output.responseCode = response.StatusCode
	output.headers = response.Header
	var err error
	output.postBody, err = ioutil.ReadAll(response.Body)
	return err
}

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
			fmt.Println(cmp.Equal(result, tt.out))

			require.Equal(t, tt.out, result, fmt.Sprintf("Expected: %v\nbut got:  %v", tt.out, result))
		})
	}
}
