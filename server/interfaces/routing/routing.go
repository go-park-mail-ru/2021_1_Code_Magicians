package routing

import (
	"net/http"
	"os"
	"pinterest/application"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/chat"
	"pinterest/interfaces/comment"
	"pinterest/interfaces/follow"
	"pinterest/interfaces/metrics"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"pinterest/interfaces/websocket"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func CreateRouter(authApp *application.AuthApp, boardInfo *board.BoardInfo, authInfo *auth.AuthInfo, profileInfo *profile.ProfileInfo,
	followInfo *follow.FollowInfo, pinInfo *pin.PinInfo, commentsInfo *comment.CommentInfo,
	websocketInfo *websocket.WebsocketInfo, notificationInfo *notification.NotificationInfo, chatInfo *chat.ChatInfo, csrfOn bool) *mux.Router {
	r := mux.NewRouter()

	r.Use(mid.PanicMid, metrics.PrometheusMiddleware)

	if csrfOn {
		csrfMid := csrf.Protect(
			[]byte(os.Getenv("CSRF_KEY")),
			csrf.Path("/"),
			csrf.Secure(false), // REMOVE IN PROD!!!!
		)
		r.Use(csrfMid)
		r.Use(mid.CSRFSettingMid)
	}

	//r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/auth/signup", mid.NoAuthMid(authInfo.HandleCreateUser, authApp)).Methods("POST")
	r.HandleFunc("/auth/login", mid.NoAuthMid(authInfo.HandleLoginUser, authApp)).Methods("POST")
	r.HandleFunc("/auth/logout", mid.AuthMid(authInfo.HandleLogoutUser, authApp)).Methods("POST")
	r.HandleFunc("/auth/check", authInfo.HandleCheckUser).Methods("GET")

	r.HandleFunc("/profile/password", mid.AuthMid(profileInfo.HandleChangePassword, authApp)).Methods("PUT")
	r.HandleFunc("/profile/edit", mid.AuthMid(profileInfo.HandleEditProfile, authApp)).Methods("PUT")
	r.HandleFunc("/profile/delete", mid.AuthMid(profileInfo.HandleDeleteProfile, authApp)).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", profileInfo.HandleGetProfile).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", profileInfo.HandleGetProfile).Methods("GET")
	r.HandleFunc("/profile", mid.AuthMid(profileInfo.HandleGetProfile, authApp)).Methods("GET")
	r.HandleFunc("/profile/avatar", mid.AuthMid(profileInfo.HandlePostAvatar, authApp)).Methods("PUT")
	r.HandleFunc("/profiles/search/{searchKey}", profileInfo.HandleGetProfilesByKeyWords).Methods("GET")

	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(followInfo.HandleFollowProfile, authApp)).Methods("POST") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(followInfo.HandleFollowProfile, authApp)).Methods("POST")
	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(followInfo.HandleUnfollowProfile, authApp)).Methods("DELETE") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(followInfo.HandleUnfollowProfile, authApp)).Methods("DELETE")
	r.HandleFunc("/followers/{id:[0-9]+}", followInfo.HandleGetFollowers).Methods("GET")
	r.HandleFunc("/following/{id:[0-9]+}", followInfo.HandleGetFollowed).Methods("GET")
	r.HandleFunc("/pins/followed", mid.AuthMid(followInfo.HandleGetFollowedPinsList, authApp)).Methods("GET")

	r.HandleFunc("/pin", mid.AuthMid(pinInfo.HandleAddPin, authApp)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", pinInfo.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pinInfo.HandleGetPinsByBoardID).Methods("GET")
	r.HandleFunc("/pin/add/{id:[0-9]+}", mid.AuthMid(pinInfo.HandleSavePin, authApp)).Methods("POST")
	r.HandleFunc("/pins/feed", pinInfo.HandlePinsFeed).Methods("GET")
	r.HandleFunc("/pins/search", pinInfo.HandleSearchPins).Methods("GET")
	r.HandleFunc("/pin/report", mid.AuthMid(pinInfo.HandleCreateReport, authApp)).Methods("POST")

	r.HandleFunc("/board", mid.AuthMid(boardInfo.HandleCreateBoard, authApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}", boardInfo.HandleGetBoardByID).Methods("GET")
	r.HandleFunc("/boards/{id:[0-9]+}", boardInfo.HandleGetBoardsByUserID).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boardInfo.HandleDelBoardByID, authApp)).Methods("DELETE")
	r.HandleFunc("/board/{id:[0-9]+}/add/{pinID:[0-9]+}", mid.AuthMid(pinInfo.HandleAddPinToBoard, authApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}/{pinID:[0-9]+}", mid.AuthMid(pinInfo.HandleDelPinByID, authApp)).Methods("DELETE")

	r.HandleFunc("/comment/{id:[0-9]+}", mid.AuthMid(commentsInfo.HandleAddComment, authApp)).Methods("POST")
	r.HandleFunc("/comments/{id:[0-9]+}", commentsInfo.HandleGetComments).Methods("GET")

	r.HandleFunc("/socket", websocketInfo.HandleConnect)
	r.HandleFunc("/notifications/read/{id:[0-9]+}", mid.AuthMid(notificationInfo.HandleReadNotification, authApp)).Methods("PUT")
	r.HandleFunc("/message/{id:[0-9]+}", mid.AuthMid(chatInfo.HandleAddMessage, authApp)).Methods("POST")
	r.HandleFunc("/message/{username}", mid.AuthMid(chatInfo.HandleAddMessage, authApp)).Methods("POST")
	r.HandleFunc("/chats/read/{id:[0-9]+}", mid.AuthMid(chatInfo.HandleReadChat, authApp)).Methods("PUT")

	if csrfOn {
		r.HandleFunc("/csrf", func(w http.ResponseWriter, r *http.Request) { // Is used only for getting csrf key
			w.WriteHeader(http.StatusCreated)
		}).Methods("GET")
	}

	return r
}
