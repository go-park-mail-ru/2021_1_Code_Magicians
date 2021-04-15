package routing

import (
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/comment"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/notifications"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateRouter(conn *pgxpool.Pool, sess *session.Session, s3BucketName string) *mux.Router {
	r := mux.NewRouter()
	r.Use(mid.PanicMid)

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

	boardsInfo := board.NewBoardInfo(boardApp)
	authInfo := auth.NewAuthInfo(userApp, cookieApp, s3App, boardApp)
	profileInfo := profile.NewProfileInfo(userApp, cookieApp, s3App)
	pinsInfo := pin.NewPinInfo(pinApp, s3App, boardApp)
	commentsInfo := comment.NewCommentInfo(commentApp, pinApp)

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
	r.HandleFunc("/pin/picture", mid.AuthMid(pinsInfo.HandleUploadPicture, cookieApp)).Methods("PUT")
	r.HandleFunc("/pin/add/{id:[0-9]+}", mid.AuthMid(pinsInfo.HandleSavePin, cookieApp)).Methods("POST")

	r.HandleFunc("/board", mid.AuthMid(boardsInfo.HandleCreateBoard, cookieApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}", boardsInfo.HandleGetBoardByID).Methods("GET")
	r.HandleFunc("/boards/{id:[0-9]+}", boardsInfo.HandleGetBoardsByUserID).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boardsInfo.HandleDelBoardByID, cookieApp)).Methods("DELETE")
	r.HandleFunc("/board/{id:[0-9]+}/add/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleAddPinToBoard, cookieApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}/{pinID:[0-9]+}", mid.AuthMid(pinsInfo.HandleDelPinByID, cookieApp)).Methods("DELETE")

	r.HandleFunc("/comment/{id:[0-9]+}", mid.AuthMid(commentsInfo.HandleAddComment, cookieApp)).Methods("POST")
	r.HandleFunc("/comments/{id:[0-9]+}", commentsInfo.HandleGetComments).Methods("GET")

	r.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024 * 1024,
			WriteBufferSize: 1024 * 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		go notifications.SendNewMsgNotifications(ws)
	})

	return r
}
