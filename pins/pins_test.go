package pins

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserPinSet_AddPinCase1(t *testing.T) {
	var testPinSet = PinsStorage {
		Storage: NewPinsSet(0),
	}
	boardID := 0
	body := strings.NewReader(
		fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "imageLink": "example/link", "description": "exampleDescription"}`, boardID),
		)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/",body )
	w := httptest.NewRecorder()

	testPinSet.Storage.AddPin(w, r)
	resResponse := w.Result()
	resBody, _ := io.ReadAll(resResponse.Body)

	require.Equal(t, http.StatusOK, resResponse.StatusCode)
	require.Equal(t, resResponse.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	require.Equal(t, string(resBody), "{pin_id: 0}")
}

func TestUserPinSet_GetPinByID(t *testing.T) {
	var testPinSet = PinsStorage {
		Storage: NewPinsSet(0),
	}
	boardID := 0
	body := strings.NewReader(
		fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "imageLink": "example/link", "description": "exampleDescription"}`, boardID),
	)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/",body )
	w := httptest.NewRecorder()

	testPinSet.Storage.AddPin(w, r)

	r = httptest.NewRequest("GET", "http://127.0.0.1:8080/pins/0", nil)
	w = httptest.NewRecorder()

	testPinSet.Storage.GetPinByID(w, r)
	resResponse := w.Result()
	resBody, _ := io.ReadAll(resResponse.Body)
	fmt.Println(resBody, "-----", resResponse.Body, "-------",resResponse)
}
