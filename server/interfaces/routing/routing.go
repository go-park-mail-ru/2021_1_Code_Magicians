package routing

import (
	"go.uber.org/zap"
	"net/http"
	"os"
	"pinterest/application"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/comment"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"time"

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

	repo := persistence.NewUserRepository(conn)
	repoPins := persistence.NewPinsRepository(conn)
	repoBoards := persistence.NewBoardsRepository(conn)
	repoComments := persistence.NewCommentsRepository(conn)

	cookieApp := application.NewCookieApp(40, 10*time.Hour)
	boardApp := application.NewBoardApp(repoBoards)
	s3App := application.NewS3App(sess, s3BucketName)
	userApp := application.NewUserApp(repo, boardApp, s3App)
	pinApp := application.NewPinApp(repoPins, boardApp, s3App)
	commentApp := application.NewCommentApp(repoComments)
	notificationsApp := application.NewNotificationApp(userApp)

	boardsInfo := board.NewBoardInfo(boardApp, zapLogger)
	authInfo := auth.NewAuthInfo(userApp, cookieApp, s3App, boardApp, notificationsApp, zapLogger)
	profileInfo := profile.NewProfileInfo(userApp, cookieApp, s3App, notificationsApp, zapLogger)
	pinsInfo := pin.NewPinInfo(pinApp, s3App, boardApp, zapLogger)
	commentsInfo := comment.NewCommentInfo(commentApp, pinApp, zapLogger)
	notificationsInfo := notification.NewNotificationInfo(notificationsApp, csrfOn, zapLogger)

	r.HandleFunc("/auth/signup", mid.NoAuthMid(authInfo.HandleCreateUser, cookieApp)).Methods("POST")
	r.HandleFunc("/auth/login", mid.NoAuthMid(authInfo.HandleLoginUser, cookieApp)).Methods("POST")
	r.HandleFunc("/auth/logout", mid.AuthMid(authInfo.HandleLogoutUser, cookieApp)).Methods("POST")
	r.HandleFunc("/auth/check", authInfo.HandleCheckUser).Methods("GET")

	r.HandleFunc("/profile/password", mid.AuthMid(profileInfo.HandleChangePassword, cookieApp)).Methods("PUT")
	r.HandleFunc("/profile/edit", mid.AuthMid(profileInfo.HandleEditProfile, cookieApp)).Methods("PUT")
	r.HandleFunc("/profile/delete", mid.AuthMid(profileInfo.HandleDeleteProfile, cookieApp)).Methods("DELETE")
	r.HandleFunc("/profile/{id:[0-9]+}", profileInfo.HandleGetProfile).Methods("GET") // Is preferred over next one
	r.HandleFunc("/profile/{username}", profileInfo.HandleGetProfile).Methods("GET")
	r.HandleFunc("/profile", mid.AuthMid(profileInfo.HandleGetProfile, cookieApp)).Methods("GET")
	r.HandleFunc("/profile/avatar", mid.AuthMid(profileInfo.HandlePostAvatar, cookieApp)).Methods("PUT")

	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(profileInfo.HandleFollowProfile, cookieApp)).Methods("POST") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(profileInfo.HandleFollowProfile, cookieApp)).Methods("POST")
	r.HandleFunc("/follow/{id:[0-9]+}", mid.AuthMid(profileInfo.HandleUnfollowProfile, cookieApp)).Methods("DELETE") // Is preferred over next one
	r.HandleFunc("/follow/{username}", mid.AuthMid(profileInfo.HandleUnfollowProfile, cookieApp)).Methods("DELETE")

	r.HandleFunc("/pin", mid.AuthMid(pinsInfo.HandleAddPin, cookieApp)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", pinsInfo.HandleGetPinByID).Methods("GET")
	r.HandleFunc("/pins/{id:[0-9]+}", pinsInfo.HandleGetPinsByBoardID).Methods("GET")
	r.HandleFunc("/pin/add/{id:[0-9]+}", mid.AuthMid(pinsInfo.HandleSavePin, cookieApp)).Methods("POST")
	r.HandleFunc("/pins/feed/{num:[0-9]+}", pinsInfo.HandlePinsFeed).Methods("GET")

	r.HandleFunc("/board", mid.AuthMid(boardsInfo.HandleCreateBoard, cookieApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}", boardsInfo.HandleGetBoardByID).Methods("GET")
	r.HandleFunc("/boards/{id:[0-9]+}", boardsInfo.HandleGetBoardsByUserID).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boardsInfo.HandleDelBoardByID, cookieApp)).Methods("DELETE")
	r.HandleFunc("/board/{id:[0-9]+}/add/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleAddPinToBoard, cookieApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleDelPinByID, cookieApp)).Methods("DELETE")

	r.HandleFunc("/comment/{id:[0-9]+}", mid.AuthMid(commentsInfo.HandleAddComment, cookieApp)).Methods("POST")
	r.HandleFunc("/comments/{id:[0-9]+}", commentsInfo.HandleGetComments).Methods("GET")

	r.HandleFunc("/notifications", notificationsInfo.HandleConnect)
	r.HandleFunc("/notifications/read/{id:[0-9]+}", mid.AuthMid(notificationsInfo.HandleReadNotification, cookieApp)).Methods("PUT")

	if csrfOn {
		r.HandleFunc("/csrf", func(w http.ResponseWriter, r *http.Request) { // Is used only for getting csrf key
			w.WriteHeader(http.StatusCreated)
		}).Methods("GET")
	}

	return r
}
