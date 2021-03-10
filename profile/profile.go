package profile

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/auth"
)

//HandleChangePassword changes profilefor user specified in request
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookieInfo, loggedIn := auth.CheckCookies(r)
	if !loggedIn {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userInput.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	currentUser := auth.Users.Users[cookieInfo.UserID]
	currentUser.Password = userInput.Password
	auth.Users.Users[cookieInfo.UserID] = currentUser
	auth.Users.Mu.Lock()

	w.WriteHeader(http.StatusCreated)
}

// HandleEditProfile edits profile specified in request
func HandleEditProfile(w http.ResponseWriter, r *http.Request) {

}

// HandleDeleteProfile deletes profile specified in request
func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {

}

// HandleGetProfile returns specified profile
func HandleGetProfile(w http.ResponseWriter, r *http.Request) {

}
