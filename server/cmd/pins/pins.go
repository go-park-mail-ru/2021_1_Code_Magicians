package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"pinterest/domain/entity"
	pinService "pinterest/services/pins"
	pinProto "pinterest/services/pins/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
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

	dbPrefix := os.Getenv("DB_PREFIX")
	if dbPrefix != "AMAZON" && dbPrefix != "LOCAL" {
		sugarLogger.Fatalf("Wrong prefix: %s , should be AMAZON or LOCAL", dbPrefix)
	}

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv(dbPrefix+"_DB_USER"), os.Getenv(dbPrefix+"_DB_PASSWORD"), os.Getenv(dbPrefix+"_DB_HOST"),
		os.Getenv(dbPrefix+"_DB_PORT"), os.Getenv(dbPrefix+"_DB_NAME"))
	conn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		sugarLogger.Fatal("Could not connect to postgres database", zap.String("error", err.Error()))
		return
	}

	fmt.Println("Successfully connected to postgres database")
	defer conn.Close()

	sess := entity.ConnectAws()

	server := grpc.NewServer()

	service := pinService.NewService(conn, sess)
	pinProto.RegisterPinsServer(server, service)

	lis, err := net.Listen("tcp", addr)

	fmt.Printf("Starting server at localhost%s\n", addr)
	err = server.Serve(lis)
	if err != nil {
		log.Fatalln("Serve auth error: ", err)
	}
}

func main() {
	runService(":8084")
}
