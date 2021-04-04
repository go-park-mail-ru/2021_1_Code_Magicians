package routing

import (
	"log"
	"net/http"
	"pinterest/auth"
	"pinterest/pins"
	"pinterest/profile"
	"pinterest/boards"

	"github.com/gorilla/mux"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

// panicMid logges error if handler errors
func panicMid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// jsonContentTypeMid adds "Content-type: application/json" to headers
func jsonContentTypeMid(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(panicMid)

	r.HandleFunc("/auth/signup", auth.NoAuthMid(auth.HandleCreateUser)).Methods("POST")
	r.HandleFunc("/auth/login", auth.NoAuthMid(auth.HandleLoginUser)).Methods("POST")
	r.HandleFunc("/auth/logout", auth.AuthMid(auth.HandleLogoutUser)).Methods("POST")
	r.HandleFunc("/auth/check", auth.HandleCheckUser).Methods("GET")

	r.HandleFunc("/profile/password", auth.AuthMid(profile.HandleChangePassword)).Methods("PUT")
	r.HandleFunc("/profile/edit", auth.AuthMid(profile.HandleEditProfile)).Methods("PUT")
	r.HandleFunc("/profile/delete", auth.AuthMid(profile.HandleDeleteProfile)).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", jsonContentTypeMid(profile.HandleGetProfile)).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", jsonContentTypeMid(profile.HandleGetProfile)).Methods("GET")
	r.HandleFunc("/profile", auth.AuthMid(jsonContentTypeMid(profile.HandleGetProfile))).Methods("GET")

	pins := &pins.PinsStorage {
		Storage: pins.NewPinsSet(),
	}
	boards := &boards.BoardsStorage {
		Storage: boards.NewBoardSet(),
	}

	r.HandleFunc("/pin", auth.AuthMid(pins.Storage.HandleAddPin)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", pins.Storage.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/pin/{id:[0-9]+}", auth.AuthMid(pins.Storage.HandleDelPinByID)).Methods("DELETE")
	r.HandleFunc("/pins/{id:[0-9]+}", auth.AuthMid(pins.Storage.HandleGetPinsByBoardID)).Methods("GET")


	r.HandleFunc("/board/", auth.AuthMid(boards.Storage.HandleAddBoard)).Methods("POST") // Will split later
	r.HandleFunc("/board/{id:[0-9]+}", auth.AuthMid(boards.Storage.HandleDelBoardByID)).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", auth.AuthMid(boards.Storage.HandleGetBoardByID)).Methods("DELETE")

	return r
}
