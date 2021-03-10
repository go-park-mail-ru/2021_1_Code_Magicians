package pins

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserPinSet_AddPinCase1(t *testing.T) {
	testPinSet := PinsStorage {
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
	assert.EqualValues(t, resResponse.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	assert.EqualValues(t, string(resBody), "{pin_id: 0}")
}

func TestUserPinSet_AddPinCase2(t *testing.T) {
	testPinSet := PinsStorage {
		Storage: NewPinsSet(0),
	}
	boardID := 0
	body := strings.NewReader(
		fmt.Sprintf(`{"boardID": %d, "title": "exampletitle", "imageLink": "example/link", "description": "exampleDescription"}`, boardID),
	)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/",body )
	w := httptest.NewRecorder()

	testPinSet.Storage.AddPin(w, r)
	testPinSet.Storage.AddPin(w, r)
	testPinSet.Storage.AddPin(w, r)
	testPinSet.Storage.AddPin(w, r)
	testPinSet.Storage.AddPin(w, r)
	for _,r := range testPinSet.Storage.userPins[0] {
		fmt.Println(r)
	}
	resResponse := w.Result()
	resBody, _ := io.ReadAll(resResponse.Body)

	require.Equal(t, http.StatusOK, resResponse.StatusCode)
	require.NotEmpty(t, w.Header().Get("Content-Type"))
	require.NotEmpty(t, resBody)

	require.Equal(t, len(testPinSet.Storage.userPins[0]), 5)
	for _,r := range testPinSet.Storage.userPins[testPinSet.Storage.userId] {
		fmt.Println(r)
	}

}
