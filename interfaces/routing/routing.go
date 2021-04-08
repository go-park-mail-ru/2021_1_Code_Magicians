package routing

import (
	"net/http"
	"pinterest/application"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/profile"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func boardHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{}"))
	// TODO /board handling
}

func CreateRouter(conn *pgx.Conn, sess *session.Session, s3BucketName string) *mux.Router {
	r := mux.NewRouter()
	r.Use(mid.PanicMid)

	repo := persistence.NewUserRepository(conn)
	authInfo := auth.AuthInfo{
		UserApp:      application.NewUserApp(repo), // TODO: mocking
		CookieApp:    application.NewCookieApp(),
		CookieLength: 40,
		Duration:     10 * time.Hour,
	}

	profileInfo := profile.ProfileInfo{
		UserApp:   authInfo.UserApp,
		CookieApp: authInfo.CookieApp,
	}

	r.HandleFunc("/auth/signup", mid.NoAuthMid(authInfo.HandleCreateUser, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/auth/login", mid.NoAuthMid(authInfo.HandleLoginUser, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/auth/logout", mid.AuthMid(authInfo.HandleLogoutUser, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/auth/check", authInfo.HandleCheckUser).Methods("GET")

	r.HandleFunc("/profile/password", mid.AuthMid(profileInfo.HandleChangePassword, profileInfo.CookieApp)).Methods("PUT")
	r.HandleFunc("/profile/edit", mid.AuthMid(profileInfo.HandleEditProfile, profileInfo.CookieApp)).Methods("PUT")
	r.HandleFunc("/profile/delete", mid.AuthMid(profileInfo.HandleDeleteProfile, profileInfo.CookieApp)).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", mid.JsonContentTypeMid(profileInfo.HandleGetProfile)).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", mid.JsonContentTypeMid(profileInfo.HandleGetProfile)).Methods("GET")
	r.HandleFunc("/profile", mid.AuthMid(mid.JsonContentTypeMid(profileInfo.HandleGetProfile), profileInfo.CookieApp)).Methods("GET")
	r.HandleFunc("/profile/avatar", mid.AuthMid(mid.AWSMid(profileInfo.HandlePostAvatar, sess, s3BucketName), profileInfo.CookieApp)).Methods("PUT")

	// pins := &pins.PinsStorage{
	// 	Storage: pin.NewPinsSet(),
	// }
	// boards := &board.BoardsStorage{
	// 	Storage: board.NewBoardSet(),
	// }

	// r.HandleFunc("/pin", mid.AuthMid(pins.Storage.HandleAddPin)).Methods("POST")
	// r.HandleFunc("/pin/{id:[0-9]+}", pins.Storage.HandleGetPinByID).Methods("GET")
	// r.HandleFunc("/pin/{id:[0-9]+}", mid.AuthMid(pins.Storage.HandleDelPinByID)).Methods("DELETE")
	// r.HandleFunc("/pins/{id:[0-9]+}", mid.AuthMid(pins.Storage.HandleGetPinsByBoardID)).Methods("GET")

	// r.HandleFunc("/board/", mid.AuthMid(boards.Storage.HandleAddBoard)).Methods("POST") // Will split later
	// r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleDelBoardByID)).Methods("GET")
	// r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleGetBoardByID)).Methods("DELETE")

	return r
}
