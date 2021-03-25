package routing

import (
	"log"
	"net/http"
	"pinterest/auth"
	"pinterest/pins"
	"pinterest/profile"

	"github.com/gorilla/mux"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /pin and /pins handling
}

// PanicMiddleware logges error if handler errors
func PanicMiddleware(next http.Handler) http.Handler {
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

func CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(PanicMiddleware)

	authNeeded := r.NewRoute().Subrouter()
	authNeeded.Use(auth.CheckAuthMiddleware)

	noAuthNeeded := r.NewRoute().Subrouter()
	noAuthNeeded.Use(auth.CheckNoAuthMiddleware)

	noAuthNeeded.HandleFunc("/auth/signup", auth.HandleCreateUser).Methods("POST")
	noAuthNeeded.HandleFunc("/auth/login", auth.HandleLoginUser).Methods("POST")
	authNeeded.HandleFunc("/auth/logout", auth.HandleLogoutUser).Methods("POST")
	r.HandleFunc("/auth/check", auth.HandleCheckUser).Methods("GET")

	authNeeded.HandleFunc("/profile/password", profile.HandleChangePassword).Methods("PUT")
	authNeeded.HandleFunc("/profile/edit", profile.HandleEditProfile).Methods("PUT")
	authNeeded.HandleFunc("/profile/delete", profile.HandleDeleteProfile).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", profile.HandleGetProfile).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", profile.HandleGetProfile).Methods("GET")
	authNeeded.HandleFunc("/profile", profile.HandleGetProfile).Methods("GET")

	pins := &pins.PinsStorage{
		Storage: pins.NewPinsSet(0),
	}

	r.HandleFunc("/pin", pins.Storage.AddPin).Methods("POST")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.GetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pins.Storage.DelPinByID).Methods("DELETE")

	r.HandleFunc("/board/", boardHandler) // Will split later

	return r
}
