package routing

import (
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/chat"
	"pinterest/interfaces/comment"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"pinterest/interfaces/websocket"
	"pinterest/usage"
)

func CreateRouter(authApp *usage.AuthApp, boardsInfo *board.BoardInfo, authInfo *auth.AuthInfo, profileInfo *profile.ProfileInfo,
	pinsInfo *pin.PinInfo, commentsInfo *comment.CommentInfo, websocketInfo *websocket.WebsocketInfo,
	notificationInfo *notification.NotificationInfo, chatInfo *chat.ChatInfo, csrfOn bool) *mux.Router {
	r := mux.NewRouter()
	r.Use(mid.PanicMid)

	if csrfOn {
		csrfMid := csrf.Protect(
			[]byte(os.Getenv("CSRF_KEY")),
			csrf.Path("/"),
			csrf.Secure(false), // REMOVE IN PROD!!!!
		)
		r.Use(csrfMid)
		r.Use(mid.CSRFSettingMid)
	}

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

	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(profileInfo.HandleFollowProfile, authApp)).Methods("POST") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(profileInfo.HandleFollowProfile, authApp)).Methods("POST")
	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(profileInfo.HandleUnfollowProfile, authApp)).Methods("DELETE") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(profileInfo.HandleUnfollowProfile, authApp)).Methods("DELETE")

	r.HandleFunc("/pin", mid.AuthMid(pinsInfo.HandleAddPin, authApp)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", pinsInfo.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pinsInfo.HandleGetPinsByBoardID).Methods("GET")
	r.HandleFunc("/pin/add/{id:[0-9]+}", mid.AuthMid(pinsInfo.HandleSavePin, authApp)).Methods("POST")
	r.HandleFunc("/pins/feed/{num:[0-9]+}", pinsInfo.HandlePinsFeed).Methods("GET")
	r.HandleFunc("/pins/search/{searchKey}", pinsInfo.HandleSearchPins).Methods("GET")

	r.HandleFunc("/board", mid.AuthMid(boardsInfo.HandleCreateBoard, authApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}", boardsInfo.HandleGetBoardByID).Methods("GET")
	r.HandleFunc("/boards/{id:[0-9]+}", boardsInfo.HandleGetBoardsByUserID).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boardsInfo.HandleDelBoardByID, authApp)).Methods("DELETE")
	r.HandleFunc("/board/{id:[0-9]+}/add/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleAddPinToBoard, authApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleDelPinByID, authApp)).Methods("DELETE")

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
