package auth

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"pinterest/usage"
	"pinterest/domain/entity"
	_ "pinterest/domain/repository"
	"pinterest/infrastructure/persistence"
	"pinterest/interfaces/routing"
	authProto "pinterest/services/auth/proto"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func runService(addr string) {
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

	fmt.Println("Successfully connected to database")
	r := routing.CreateRouter(conn, sess, os.Getenv("BUCKET_NAME"), os.Getenv("CSRF_ON") == "true")

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

	server := grpc.NewServer()
	userAuthRepository := persistence.NewUserRepository(conn)
	cookieApp := usage.NewCookieApp(40, 10*time.Hour)

	service := usage.NewAuthApp(userAuthRepository, cookieApp)
	authProto.RegisterAuthServer(server, service)

	lis, err := net.Listen("tcp", addr)

	err = server.Serve(lis)
	if err != nil {
		log.Fatalln("Serve auth error: ", err)
	}
}

func main() {
	runService(":8081")
}
