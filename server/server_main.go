package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"pinterest/domain/entity"
	"pinterest/interfaces/auth"
	"pinterest/interfaces/board"
	"pinterest/interfaces/chat"
	"pinterest/interfaces/comment"
	"pinterest/interfaces/notification"
	"pinterest/interfaces/pin"
	"pinterest/interfaces/profile"
	"pinterest/interfaces/routing"
	"pinterest/interfaces/websocket"
	protoUser "pinterest/services/user/proto"
	protoAuth "pinterest/services/auth/proto"
	protoComments "pinterest/services/comments/proto"
	protoPins "pinterest/services/pins/proto"
	"pinterest/usage"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
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
	// TODO: check if all needed variables are present

	dbPrefix := os.Getenv("DB_PREFIX")
	if dbPrefix != "AMAZON" && dbPrefix != "LOCAL" {
		sugarLogger.Fatalf("Wrong prefix: %s , should be AMAZON or LOCAL", dbPrefix)
	}

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv(dbPrefix+"_DB_USER"), os.Getenv(dbPrefix+"_DB_PASSWORD"), os.Getenv(dbPrefix+"_DB_HOST"),
		os.Getenv(dbPrefix+"_DB_PORT"), os.Getenv(dbPrefix+"_DB_NAME"))
	conn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		sugarLogger.Fatal("Could not connect to database", zap.String("error", err.Error()))
		return
	}
	defer conn.Close()
	sess := entity.ConnectAws()
	// TODO divide file

	//
	//var kacp = keepalive.ClientParameters{
	//	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	//	Timeout:             time.Second,      // wait 1 second for ping back
	//	PermitWithoutStream: true,             // send pings even without active streams
	//}

	sessionUser, err := grpc.Dial("127.0.0.1:8082", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for User service")
	}
	defer sessionUser.Close()

	sessionAuth, err := grpc.Dial("127.0.0.1:8083", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Auth service")
	}
	defer sessionAuth.Close()

	sessionPins, err := grpc.Dial("127.0.0.1:8084", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Pins service")
	}
	defer sessionPins.Close()

	sessionComments, err := grpc.Dial("127.0.0.1:8085", grpc.WithInsecure())
	if err != nil {
		sugarLogger.Fatal("Can not create session for Comments service")
	}
	defer sessionComments.Close()

	repoUser := protoUser.NewUserClient(sessionUser)
	repoAuth := protoAuth.NewAuthClient(sessionAuth)
	repoPins := protoPins.NewPinsClient(sessionPins)
	repoComments := protoComments.NewCommentsClient(sessionComments)

	cookieApp := usage.NewCookieApp(40, 10*time.Hour)
	boardApp := usage.NewBoardApp(repoPins)
	s3App := usage.NewS3App(sess, os.Getenv("BUCKET_NAME"))
	userApp := usage.NewUserApp(repoUser, boardApp)
	authApp := usage.NewAuthApp(repoAuth, userApp, cookieApp)
	pinApp := usage.NewPinApp(repoPins, boardApp)
	commentApp := usage.NewCommentApp(repoComments)
	websocketApp := usage.NewWebsocketApp(userApp)
	notificationApp := usage.NewNotificationApp(userApp, websocketApp)
	chatApp := usage.NewChatApp(userApp, websocketApp)

	boardsInfo := board.NewBoardInfo(boardApp, logger)
	authInfo := auth.NewAuthInfo(authApp, userApp, cookieApp, s3App, boardApp, websocketApp, logger)
	profileInfo := profile.NewProfileInfo(userApp, authApp, cookieApp, s3App, notificationApp, logger)
	pinsInfo := pin.NewPinInfo(pinApp, s3App, boardApp, logger)
	commentsInfo := comment.NewCommentInfo(commentApp, pinApp, logger)
	websocketInfo := websocket.NewWebsocketInfo(notificationApp, chatApp, websocketApp, os.Getenv("CSRF_ON") == "true", logger)
	notificationInfo := notification.NewNotificationInfo(notificationApp, logger)
	chatInfo := chat.NewChatnfo(chatApp, userApp, logger)
	// TODO divide file

	fmt.Println("Successfully connected to database")
	r := routing.CreateRouter(authApp, boardsInfo, authInfo, profileInfo, pinsInfo, commentsInfo,
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
