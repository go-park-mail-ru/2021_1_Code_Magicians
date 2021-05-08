package routing

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net/http"
	"os"
	"pinterest/usage"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/chat"
	"pinterest/interfaces/comment"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"pinterest/interfaces/websocket"
	protoUser "pinterest/services/user/proto"
	"time"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateRouter(conn *pgxpool.Pool, sess *session.Session, s3BucketName string, csrfOn bool) *mux.Router {
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
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping back
		PermitWithoutStream: true,             // send pings even without active streams
	}

	sessionUser, _ := grpc.Dial("127.0.0.1:8082", grpc.WithInsecure(),grpc.WithKeepaliveParams(kacp))
	defer sessionUser.Close()
	repo := protoUser.NewUserClient(sessionUser)
	repo1 := persistence.NewUserRepository(conn)
	repoPins := persistence.NewPinsRepository(conn)
	repoBoards := persistence.NewBoardsRepository(conn)
	repoComments := persistence.NewCommentsRepository(conn)

	cookieApp := usage.NewCookieApp(40, 10*time.Hour)
	authApp := usage.NewAuthApp(repo1, cookieApp)
	boardApp := usage.NewBoardApp(repoBoards)
	s3App := usage.NewS3App(sess, s3BucketName)
	userApp := usage.NewUserApp(repo, boardApp, s3App)
	pinApp := usage.NewPinApp(repoPins, boardApp, s3App)
	commentApp := usage.NewCommentApp(repoComments)
	websocketApp := usage.NewWebsocketApp(userApp)
	notificationApp := usage.NewNotificationApp(userApp, websocketApp)
	chatApp := usage.NewChatApp(userApp, websocketApp)

	boardsInfo := board.NewBoardInfo(boardApp, zapLogger)
	authInfo := auth.NewAuthInfo(authApp, userApp, cookieApp, s3App, boardApp, websocketApp, zapLogger)
	profileInfo := profile.NewProfileInfo(userApp, cookieApp, s3App, notificationApp, zapLogger)
	pinsInfo := pin.NewPinInfo(pinApp, s3App, boardApp, zapLogger)
	commentsInfo := comment.NewCommentInfo(commentApp, pinApp, zapLogger)
	websocketInfo := websocket.NewWebsocketInfo(notificationApp, chatApp, websocketApp, csrfOn, zapLogger)
	notificationInfo := notification.NewNotificationInfo(notificationApp, zapLogger)
	chatInfo := chat.NewChatnfo(chatApp, userApp, zapLogger)

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
