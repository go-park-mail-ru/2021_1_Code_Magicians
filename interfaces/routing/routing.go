package routing

import (
	"pinterest/application"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func CreateRouter(conn *pgx.Conn) *mux.Router {
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

 	repoPins := persistence.NewPinsRepository(conn)
	pinsInfo := pin.PinInfo{
		PinApp: application.NewPinApp(repoPins),
	}

	//repoBoards := persistence.NewBoardsRepository(conn)

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
	r.HandleFunc("/profile/avatar", mid.AuthMid(profileInfo.HandlePostAvatar, profileInfo.CookieApp)).Methods("PUT")

	r.HandleFunc("/pin", mid.AuthMid(pinsInfo.HandleAddPin, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", mid.JsonContentTypeMid(pinsInfo.HandleGetPinByID)).Methods("GET")
	r.HandleFunc("/pin/{id:[0-9]+}", mid.AuthMid(pinsInfo.HandleDelPinByID, authInfo.CookieApp)).Methods("DELETE")
	r.HandleFunc("/pins/{id:[0-9]+}", mid.JsonContentTypeMid(pinsInfo.HandleGetPinsByBoardID)).Methods("GET")

	// r.HandleFunc("/board/", mid.AuthMid(boards.Storage.HandleAddBoard)).Methods("POST") // Will split later
	// r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleDelBoardByID)).Methods("GET")
	// r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boards.Storage.HandleGetBoardByID)).Methods("DELETE")

	return r
}
