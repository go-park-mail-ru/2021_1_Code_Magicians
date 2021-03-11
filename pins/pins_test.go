package pins

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testPinSet = PinsStorage{
	Storage: NewPinsSet(0),
}

func TestUserPinSet_AddPin(t *testing.T) {
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

	require.Equal(t, http.StatusOK, resResponse.StatusCode)
	require.Equal(t, resResponse.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	require.Equal(t, string(resBody), "{pin_id: 1}")
	require.Equal(t, len(testPinSet.Storage.userPins[0]), 2)
}

func TestUserPinSet_AddPinBadData(t *testing.T) {
	body := strings.NewReader(
		fmt.Sprintf(`{"bad data"}`),
	)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8080/pin/", body)
	w := httptest.NewRecorder()
	testPinSet.Storage.AddPin(w, r)

	resResponse := w.Result()

	require.Equal(t, http.StatusInternalServerError, resResponse.StatusCode)
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
