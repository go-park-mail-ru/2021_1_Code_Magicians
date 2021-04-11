package routing

import (
	"pinterest/application"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/comment"
	mid "pinterest/interfaces/middleware"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func CreateRouter(conn *pgx.Conn, sess *session.Session, s3BucketName string) *mux.Router {
	r := mux.NewRouter()
	r.Use(mid.PanicMid)

	repo := persistence.NewUserRepository(conn)
	repoPins := persistence.NewPinsRepository(conn)
	repoBoards := persistence.NewBoardsRepository(conn)

	boardsInfo := board.BoardInfo{
		BoardApp: application.NewBoardApp(repoBoards),
	}
	authInfo := auth.AuthInfo{
		UserApp:      application.NewUserApp(repo),
		CookieApp:    application.NewCookieApp(),
		BoardApp:     boardsInfo.BoardApp,
		CookieLength: 40,
		Duration:     10 * time.Hour,
	}

	profileInfo := profile.ProfileInfo{
		UserApp:   authInfo.UserApp,
		CookieApp: authInfo.CookieApp,
		S3App:     application.NewS3App(sess, s3BucketName),
	}

	pinsInfo := pin.PinInfo{
		PinApp: application.NewPinApp(repoPins),
		S3App:  profileInfo.S3App,
	}

	repoComments := persistence.NewCommentsRepository(conn)
	commentsInfo := comment.CommentInfo{
		PinApp:     pinsInfo.PinApp,
		CommentApp: application.NewCommentApp(repoComments),
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
	r.HandleFunc("/profile/avatar", mid.AuthMid(profileInfo.HandlePostAvatar, profileInfo.CookieApp)).Methods("PUT")

	r.HandleFunc("/follow/{id:[0-9]+}", profileInfo.HandleFollowProfile).Methods("POST") // Is preferred over next one
	r.HandleFunc("/follow/{username}", profileInfo.HandleFollowProfile).Methods("POST")
	r.HandleFunc("/follow/{id:[0-9]+}", profileInfo.HandleUnfollowProfile).Methods("DELETE") // Is preferred over next one
	r.HandleFunc("/follow/{username}", profileInfo.HandleUnfollowProfile).Methods("DELETE")

	r.HandleFunc("/pin", mid.AuthMid(pinsInfo.HandleAddPin, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/pin/{id:[0-9]+}", mid.JsonContentTypeMid(pinsInfo.HandleGetPinByID)).Methods("GET")
	r.HandleFunc("/pin/{id:[0-9]+}", mid.AuthMid(pinsInfo.HandleDelPinByID, authInfo.CookieApp)).Methods("DELETE")
	r.HandleFunc("/pins/{id:[0-9]+}", mid.JsonContentTypeMid(pinsInfo.HandleGetPinsByBoardID)).Methods("GET")
	r.HandleFunc("/pin/picture", mid.AuthMid(pinsInfo.HandleUploadPicture, authInfo.CookieApp)).Methods("PUT")

	r.HandleFunc("/board", mid.AuthMid(boardsInfo.HandleAddBoard, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/board/{id:[0-9]+}", mid.JsonContentTypeMid(boardsInfo.HandleGetBoardByID)).Methods("GET")
	r.HandleFunc("/boards/{id:[0-9]+}", mid.JsonContentTypeMid(boardsInfo.HandleGetBoardsByUserID)).Methods("GET")
	r.HandleFunc("/board/{id:[0-9]+}", mid.AuthMid(boardsInfo.HandleDelBoardByID, authInfo.CookieApp)).Methods("DELETE")

	r.HandleFunc("/comment/{id:[0-9]+}", mid.AuthMid(commentsInfo.HandleAddComment, authInfo.CookieApp)).Methods("POST")
	r.HandleFunc("/comments/{id:[0-9]+}", mid.JsonContentTypeMid(commentsInfo.HandleGetComments)).Methods("GET")

	return r
}
