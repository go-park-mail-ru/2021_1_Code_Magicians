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

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func CreateRouter(authApp *application.AuthApp, boardInfo *board.BoardInfo, authInfo *auth.AuthInfo, profileInfo *profile.ProfileInfo,
	followInfo *follow.FollowInfo, pinInfo *pin.PinInfo, commentsInfo *comment.CommentInfo,
	websocketInfo *websocket.WebsocketInfo, notificationInfo *notification.NotificationInfo, chatInfo *chat.ChatInfo, csrfOn bool, httpOn bool) *mux.Router {
	r := mux.NewRouter()

	r.Use(mid.PanicMid, metrics.PrometheusMiddleware)

	if csrfOn {
		csrfMid := csrf.Protect(
			[]byte(os.Getenv("CSRF_KEY")),
			csrf.Path("/"),
			csrf.Secure(httpOn), // REMOVE IN PROD!!!!
		)
		r.Use(csrfMid)
		r.Use(mid.CSRFSettingMid)
	}

	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/api/auth/signup", mid.NoAuthMid(authInfo.HandleCreateUser, authApp)).Methods("POST")
	r.HandleFunc("/api/auth/login", mid.NoAuthMid(authInfo.HandleLoginUser, authApp)).Methods("POST")
	r.HandleFunc("/api/auth/logout", mid.AuthMid(authInfo.HandleLogoutUser, authApp)).Methods("POST")
	r.HandleFunc("/api/auth/check", authInfo.HandleCheckUser).Methods("GET")
	r.HandleFunc("/api/vk_token/signup", mid.NoAuthMid(authInfo.HandleCreateUserWithVK, authApp)).Methods("POST")
	r.HandleFunc("/api/vk_token/login", mid.NoAuthMid(authInfo.HandleCheckVkToken, authApp)).Methods("POST")
	r.HandleFunc("/api/vk_token/add", mid.AuthMid(authInfo.HandleAddVkToken, authApp)).Methods("POST")

	r.HandleFunc("/api/profile/password", mid.AuthMid(profileInfo.HandleChangePassword, authApp)).Methods("PUT")
	r.HandleFunc("/api/profile/edit", mid.AuthMid(profileInfo.HandleEditProfile, authApp)).Methods("PUT")
	r.HandleFunc("/api/profile/delete", mid.AuthMid(profileInfo.HandleDeleteProfile, authApp)).Methods("DELETE")
	r.HandleFunc("/api/profile/{id:[0-9]+}", profileInfo.HandleGetProfile).Methods("GET") // Is preferred over next one
	r.HandleFunc("/api/profile/{username}", profileInfo.HandleGetProfile).Methods("GET")
	r.HandleFunc("/api/profile", mid.AuthMid(profileInfo.HandleGetProfile, authApp)).Methods("GET")
	r.HandleFunc("/api/profile/avatar", mid.AuthMid(profileInfo.HandlePostAvatar, authApp)).Methods("PUT")
	r.HandleFunc("/api/profiles/search/{searchKey}", profileInfo.HandleGetProfilesByKeyWords).Methods("GET")

	r.HandleFunc("/api/follow/{id:[0-9]+}", mid.AuthMid(followInfo.HandleFollowProfile, authApp)).Methods("POST") // Is preferred over next one
	r.HandleFunc("/api/follow/{username}", mid.AuthMid(followInfo.HandleFollowProfile, authApp)).Methods("POST")
	r.HandleFunc("/api/follow/{id:[0-9]+}", mid.AuthMid(followInfo.HandleUnfollowProfile, authApp)).Methods("DELETE") // Is preferred over next one
	r.HandleFunc("/api/follow/{username}", mid.AuthMid(followInfo.HandleUnfollowProfile, authApp)).Methods("DELETE")
	r.HandleFunc("/api/followers/{id:[0-9]+}", followInfo.HandleGetFollowers).Methods("GET")
	r.HandleFunc("/api/following/{id:[0-9]+}", followInfo.HandleGetFollowed).Methods("GET")
	r.HandleFunc("/api/pins/followed", mid.AuthMid(followInfo.HandleGetFollowedPinsList, authApp)).Methods("GET")

	r.HandleFunc("/api/pin", mid.AuthMid(pinInfo.HandleAddPin, authApp)).Methods("POST")
	r.HandleFunc("/api/pin/{id:[0-9]+}", pinInfo.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/api/pins/{id:[0-9]+}", pinInfo.HandleGetPinsByBoardID).Methods("GET")
	r.HandleFunc("/api/pin/add/{id:[0-9]+}", mid.AuthMid(pinInfo.HandleSavePin, authApp)).Methods("POST")
	r.HandleFunc("/api/pins/feed", pinInfo.HandlePinsFeed).Methods("GET")
	r.HandleFunc("/api/pins/search", pinInfo.HandleSearchPins).Methods("GET")
	r.HandleFunc("/api/pin/report", mid.AuthMid(pinInfo.HandleCreateReport, authApp)).Methods("POST")

	r.HandleFunc("/api/board", mid.AuthMid(boardInfo.HandleCreateBoard, authApp)).Methods("POST")
	r.HandleFunc("/api/board/{id:[0-9]+}", boardInfo.HandleGetBoardByID).Methods("GET")
	r.HandleFunc("/api/boards/{id:[0-9]+}", boardInfo.HandleGetBoardsByUserID).Methods("GET")
	r.HandleFunc("/api/board/{id:[0-9]+}", mid.AuthMid(boardInfo.HandleDelBoardByID, authApp)).Methods("DELETE")
	r.HandleFunc("/api/board/{id:[0-9]+}/add/{pinID:[0-9]+}", mid.AuthMid(pinInfo.HandleAddPinToBoard, authApp)).Methods("POST")
	r.HandleFunc("/api/board/{id:[0-9]+}/{pinID:[0-9]+}", mid.AuthMid(pinInfo.HandleDelPinByID, authApp)).Methods("DELETE")

	r.HandleFunc("/api/comment/{id:[0-9]+}", mid.AuthMid(commentsInfo.HandleAddComment, authApp)).Methods("POST")
	r.HandleFunc("/api/comments/{id:[0-9]+}", commentsInfo.HandleGetComments).Methods("GET")

	r.HandleFunc("/socket", websocketInfo.HandleConnect)
	r.HandleFunc("/api/notifications/read/{id:[0-9]+}", mid.AuthMid(notificationInfo.HandleReadNotification, authApp)).Methods("PUT")
	r.HandleFunc("/api/message/{id:[0-9]+}", mid.AuthMid(chatInfo.HandleAddMessage, authApp)).Methods("POST")
	r.HandleFunc("/api/message/{username}", mid.AuthMid(chatInfo.HandleAddMessage, authApp)).Methods("POST")
	r.HandleFunc("/api/chats/read/{id:[0-9]+}", mid.AuthMid(chatInfo.HandleReadChat, authApp)).Methods("PUT")

	if csrfOn {
		r.HandleFunc("/api/csrf", func(w http.ResponseWriter, r *http.Request) { // Is used only for getting csrf key
			w.WriteHeader(http.StatusCreated)
		}).Methods("GET")
	}

	return r
}
