package profile

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"pinterest/auth"
	"strconv"

	"github.com/gorilla/mux"
)

//HandleChangePassword changes profilefor user specified in request
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookieInfo, loggedIn := auth.CheckCookies(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserIO)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// HandleEditProfile edits profile of current user
func HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookieInfo, loggedIn := auth.CheckCookies(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserIO)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userInput.Username != "" || userInput.Password != "" { // username is unchangeable, password is changed through a different function
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	currentUser := auth.Users.Users[cookieInfo.UserID]
	if userInput.FirstName != "" {
		currentUser.FirstName = userInput.FirstName
	}
	if userInput.LastName != "" {
		currentUser.LastName = userInput.LastName
	}
	if userInput.Email != "" {
		currentUser.Email = userInput.Email
	}
	if userInput.Avatar != "" {
		currentUser.Avatar = userInput.Avatar
	}
	auth.Users.Users[cookieInfo.UserID] = currentUser
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// HandleDeleteProfile deletes profile of current user
func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookieInfo, loggedIn := auth.CheckCookies(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	auth.HandleLogoutUser(w, r) // User is logged out before profile deletion, for safety reasons

	auth.Users.Mu.Lock()
	delete(auth.Users.Users, cookieInfo.UserID)
	auth.Users.Mu.Unlock()
}

// HandleGetProfile returns specified profile
func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, passedID := vars["id"]
	id, _ := strconv.Atoi(idStr)

	if !passedID {
		username, passedUsername := vars["username"]
		if !passedUsername {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var foundUsername bool
		id, foundUsername = auth.FindUser(username)
		if !foundUsername {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	auth.Users.Mu.Lock()
	user := auth.Users.Users[id]
	auth.Users.Mu.Unlock()

	userOutput := auth.UserIO{
		Username: user.Username,
		// Password is ommitted on purpose
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Avatar:    user.Avatar,
	}

	responseBody, err := json.Marshal(userOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)
	return
}
