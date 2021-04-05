package profile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pinterest/auth"
	"strconv"

	"github.com/gorilla/mux"
)

//HandleChangePassword changes profilefor user specified in request
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("userID").(int)

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserPassChangeInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, err := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	currentUser := auth.Users.Users[userID]
	currentUser.Password = userInput.Password
	auth.Users.Users[userID] = currentUser
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// HandleEditProfile edits profile of current user
func HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("userID").(int)

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserEditInput)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid, err := userInput.Validate()
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	newUser := auth.Users.Users[userID] // newUser is a copy which we can modify freely
	auth.Users.Mu.Unlock()

	err = newUser.UpdateFrom(userInput)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userInput.Username != "" { // username uniqueness check
		_, alreadyExists := auth.FindUser(newUser.Username)
		if alreadyExists {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	auth.Users.Mu.Lock()
	auth.Users.Users[userID] = newUser
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteProfile deletes profile of current user
func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("userID").(int)

	auth.HandleLogoutUser(w, r) // User is logged out before profile deletion, for safety reasons

	auth.Users.Mu.Lock()
	delete(auth.Users.Users, userID)
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetProfile returns specified profile
func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, passedID := vars["id"]
	id, _ := strconv.Atoi(idStr)

	if !passedID { // Id was not passed
		username, passedUsername := vars["username"]
		var foundUsername bool
		id, foundUsername = auth.FindUser(username)

		if !passedUsername { // Username was also not passed
			userID := r.Context().Value("userID")
			if userID == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			id = userID.(int)
		} else if !foundUsername {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	auth.Users.Mu.Lock()
	user := auth.Users.Users[id]
	auth.Users.Mu.Unlock()

	var userOutput auth.UserOutput
	userOutput.FillFromUser(&user)
	userOutput.Password = "" // Password is ommitted on purpose

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
