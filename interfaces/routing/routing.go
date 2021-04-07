package routing

import (
	"net/http"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"

	"github.com/gorilla/mux"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

func CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(mid.PanicMid)

	r.HandleFunc("/auth/signup", mid.NoAuthMid(auth.HandleCreateUser)).Methods("POST")
	r.HandleFunc("/auth/login", mid.NoAuthMid(auth.HandleLoginUser)).Methods("POST")
	r.HandleFunc("/auth/logout", mid.AuthMid(auth.HandleLogoutUser)).Methods("POST")
	r.HandleFunc("/auth/check", auth.HandleCheckUser).Methods("GET")

	r.HandleFunc("/profile/password", mid.AuthMid(profile.HandleChangePassword)).Methods("PUT")
	r.HandleFunc("/profile/edit", mid.AuthMid(profile.HandleEditProfile)).Methods("PUT")
	r.HandleFunc("/profile/delete", mid.AuthMid(profile.HandleDeleteProfile)).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", mid.JsonContentTypeMid(profile.HandleGetProfile)).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", mid.JsonContentTypeMid(profile.HandleGetProfile)).Methods("GET")
	r.HandleFunc("/profile", mid.AuthMid(mid.JsonContentTypeMid(profile.HandleGetProfile))).Methods("GET")

	pins := &pins.PinsStorage{
		Storage: pin.NewPinsSet(),
	}
	boards := &board.BoardsStorage{
		Storage: board.NewBoardSet(),
	}

	r.HandleFunc("/pin", mid.AuthMid(pins.Storage.HandleAddPin)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", pins.Storage.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/pin/{id:[0-9]+}", mid.AuthMid(pins.Storage.HandleDelPinByID)).Methods("DELETE")
	r.HandleFunc("/pins/{id:[0-9]+}", mid.AuthMid(pins.Storage.HandleGetPinsByBoardID)).Methods("GET")

	r.HandleFunc("/board/", mid.AuthMid(boards.Storage.HandleAddBoard)).Methods("POST") // Will split later
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleDelBoardByID)).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleGetBoardByID)).Methods("DELETE")

	return r
}
