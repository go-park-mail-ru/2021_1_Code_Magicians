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

	userID := r.Context().Value("userID").(int)

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserIO)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userInput.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if *userInput.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	currentUser := auth.Users.Users[userID]
	currentUser.Password = *userInput.Password
	auth.Users.Users[userID] = currentUser
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// HandleEditProfile edits profile of current user
func HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("userID").(int)

	body, _ := ioutil.ReadAll(r.Body)

	userInput := new(auth.UserIO)
	err := json.Unmarshal(body, userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userInput.Password != nil { // username is unchangeable, password is changed through a different function
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	newUser := auth.Users.Users[userID] // newUser is a copy which we can modify freely
	auth.Users.Mu.Unlock()

	userInput.UpdateUser(&newUser)

	if userInput.Username != nil {
		if newUser.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, alreadyExists := auth.FindUser(newUser.Username)
		if alreadyExists {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	if userInput.Email != nil && newUser.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth.Users.Mu.Lock()
	auth.Users.Users[userID] = newUser
	auth.Users.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// HandleDeleteProfile deletes profile of current user
func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID := r.Context().Value("userID").(int)

	auth.HandleLogoutUser(w, r) // User is logged out before profile deletion, for safety reasons

	auth.Users.Mu.Lock()
	delete(auth.Users.Users, userID)
	auth.Users.Mu.Unlock()
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

	var userOutput auth.UserIO
	user.Password = "" // Password is ommitted on purpose
	userOutput.FillFromUser(&user)

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
