package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	authService "pinterest/services/auth"
	authProto "pinterest/services/auth/proto"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/tarantool/go-tarantool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func runService(addr string) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	sugarLogger := logger.Sugar()

	err := godotenv.Load(".env")
	if err != nil {
		sugarLogger.Fatal("Could not load .env file", zap.String("error", err.Error()))
	}

	err = godotenv.Load("passwords.env")
	if err != nil {
		sugarLogger.Fatal("Could not load passwords.env file", zap.String("error", err.Error()))
	}

	err = godotenv.Load("s3.env")
	if err != nil {
		sugarLogger.Fatal("Could not load s3.env file", zap.String("error", err.Error()))
	}

	err = godotenv.Load("docker_vars.env")
	if err != nil {
		sugarLogger.Fatal("Could not load docker_vars.env file", zap.String("error", err.Error()))
	}

	dbPrefix := os.Getenv("DB_PREFIX")
	if dbPrefix != "AMAZON" && dbPrefix != "LOCAL" {
		sugarLogger.Fatalf("Wrong prefix: %s , should be AMAZON or LOCAL", dbPrefix)
	}

	postgresConnectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv(dbPrefix+"_DB_USER"), os.Getenv(dbPrefix+"_DB_PASSWORD"), os.Getenv(dbPrefix+"_DB_HOST"),
		os.Getenv(dbPrefix+"_DB_PORT"), os.Getenv(dbPrefix+"_DB_NAME"))
	postgresConn, err := pgxpool.Connect(context.Background(), postgresConnectionString)
	if err != nil {
		sugarLogger.Fatal("Could not connect to postgres database", zap.String("error", err.Error()))
		return
	}

	fmt.Println("Successfully connected to postgres database")
	defer postgresConn.Close()

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

	server := grpc.NewServer()

	service := authService.NewService(postgresConn, tarantoolConn)
	authProto.RegisterAuthServer(server, service)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Listen auth error: ", err)
	}

	fmt.Printf("Starting server at localhost%s\n", addr)
	err = server.Serve(lis)
	if err != nil {
		log.Fatalln("Serve auth error: ", err)
	}
}

func main() {
	runService(":8083")
}
