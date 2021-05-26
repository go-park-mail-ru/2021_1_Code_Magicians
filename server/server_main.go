package main

import (
	"fmt"
	"net/http"
	"os"
	"pinterest/application"
	"pinterest/domain/entity"
	"pinterest/infrastructure/persistance"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/chat"
	"pinterest/interfaces/comment"
	"pinterest/interfaces/follow"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"pinterest/interfaces/routing"
	"pinterest/interfaces/websocket"
	protoAuth "pinterest/services/auth/proto"
	protoComments "pinterest/services/comments/proto"
	protoPins "pinterest/services/pins/proto"
	protoUser "pinterest/services/user/proto"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/tarantool/go-tarantool"
)

func runServer(addr string) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	sugarLogger := logger.Sugar()

	err := godotenv.Load(".env")
	if err != nil {
		sugarLogger.Fatal("Could not load .env file", zap.String("error", err.Error()))
	}

	err = godotenv.Load("s3.env")
	if err != nil {
		sugarLogger.Fatal("Could not load s3.env file", zap.String("error", err.Error()))
	}

	err = godotenv.Load("docker_vars.env")
	if err != nil {
		sugarLogger.Fatal("Could not load docker_vars.env file", zap.String("error", err.Error()))
	}

	dockerStatus := os.Getenv("CONTAINER_PREFIX")
	if dockerStatus != "DOCKER" && dockerStatus != "LOCALHOST" {
		sugarLogger.Fatalf("Wrong prefix: %s , should be DOCKER or LOCALHOST", dockerStatus)
	}

	tarantoolConn, err := tarantool.Connect(os.Getenv(dockerStatus+"_TARANTOOL_PREFIX")+":3301", tarantool.Opts{
		User: os.Getenv("TARANTOOL_USER"),
		Pass: os.Getenv("TARANTOOL_PASSWORD"),
	})
	if err != nil {
		sugarLogger.Fatal("Could not connect to tarantool database", zap.String("error", err.Error()))
	}

	fmt.Println("Successfully connected to tarantool database")
	defer tarantoolConn.Close()

	sess := entity.ConnectAws()
	// TODO divide file

	sessionUser, err := grpc.Dial(os.Getenv(dockerStatus+"_USER_PREFIX")+":8082", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for User service")
	}
	defer sessionUser.Close()

	sessionAuth, err := grpc.Dial(os.Getenv(dockerStatus+"_AUTH_PREFIX")+":8083", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Auth service")
	}
	defer sessionAuth.Close()

	sessionPins, err := grpc.Dial(os.Getenv(dockerStatus+"_PINS_PREFIX")+":8084", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Pins service")
	}
	defer sessionPins.Close()

	sessionComments, err := grpc.Dial(os.Getenv(dockerStatus+"_COMMENTS_PREFIX")+":8085", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Comments service")
	}
	defer sessionComments.Close()

	repoUser := protoUser.NewUserClient(sessionUser)
	repoAuth := protoAuth.NewAuthClient(sessionAuth)
	repoPins := protoPins.NewPinsClient(sessionPins)
	repoComments := protoComments.NewCommentsClient(sessionComments)
	repoNotification := persistance.NewNotificationRepository(tarantoolConn)
	repoChat := persistance.NewChatRepository(tarantoolConn)

	cookieApp := application.NewCookieApp(repoAuth, 40, 10*time.Hour)
	boardApp := application.NewBoardApp(repoPins)
	s3App := application.NewS3App(sess, os.Getenv("BUCKET_NAME"))
	userApp := application.NewUserApp(repoUser, boardApp)
	authApp := application.NewAuthApp(repoAuth, userApp, cookieApp)
	pinApp := application.NewPinApp(repoPins, boardApp)
	followApp := application.NewFollowApp(repoUser, pinApp)
	commentApp := application.NewCommentApp(repoComments)
	websocketApp := application.NewWebsocketApp(userApp)
	notificationApp := application.NewNotificationApp(repoNotification, userApp, websocketApp)
	chatApp := application.NewChatApp(repoChat, userApp, websocketApp)

	boardInfo := board.NewBoardInfo(boardApp, logger)
	authInfo := auth.NewAuthInfo(userApp, authApp, cookieApp, s3App, boardApp, websocketApp, logger)
	profileInfo := profile.NewProfileInfo(userApp, authApp, cookieApp, followApp, s3App, notificationApp, logger)
	followInfo := follow.NewFollowInfo(userApp, followApp, notificationApp, logger)
	pinInfo := pin.NewPinInfo(pinApp, s3App, boardApp, logger)
	commentsInfo := comment.NewCommentInfo(commentApp, pinApp, logger)
	websocketInfo := websocket.NewWebsocketInfo(notificationApp, chatApp, websocketApp, os.Getenv("CSRF_ON") == "true", logger)
	notificationInfo := notification.NewNotificationInfo(notificationApp, logger)
	chatInfo := chat.NewChatnfo(chatApp, userApp, logger)
	// TODO divide file

	r := routing.CreateRouter(authApp, boardInfo, authInfo, profileInfo, followInfo, pinInfo, commentsInfo,
		websocketInfo, notificationInfo, chatInfo, os.Getenv("CSRF_ON") == "true")

	allowedOrigins := make([]string, 3) // If needed, replace 3 with number of needed origins
	switch os.Getenv("HTTPS_ON") {
	case "true":
		allowedOrigins = append(allowedOrigins, "https://pinter-best.com:8081", "https://pinter-best.com", "https://127.0.0.1:8081")
	case "false":
		allowedOrigins = append(allowedOrigins, "http://pinter-best.com:8081", "http://pinter-best.com", "http://127.0.0.1:8081")
	default:
		sugarLogger.Fatal("HTTPS_ON variable is not set")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	})

	handler := c.Handler(r)
	fmt.Printf("Starting server at localhost%s\n", addr)

	switch os.Getenv("HTTPS_ON") {
	case "true":
		sugarLogger.Fatal(http.ListenAndServeTLS(addr, "cert.pem", "key.pem", handler))
	case "false":
		sugarLogger.Fatal(http.ListenAndServe(addr, handler))
	}
}

func main() {
	runServer(":8080")
}
