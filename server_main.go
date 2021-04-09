package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pinterest/interfaces/routing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

// connectAws returns session that can be used to connect to Amazon Web Service
func connectAws() *session.Session {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	myRegion := os.Getenv("AWS_REGION")
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(myRegion),
			Credentials: credentials.NewStaticCredentials(
				accessKeyID,
				secretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}
	return sess
}

func runServer(addr string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
		fmt.Println("Could not load .env file")
	}
	err = godotenv.Load("s3.env")
	if err != nil {
		log.Println(err)
		fmt.Println("Could not load s3.env file")
	}
	// TODO: check if all needed variables are present

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("LOCAL_DB_USER"), os.Getenv("LOCAL_DB_PASSWORD"), os.Getenv("LOCAL_DB_HOST"),
		os.Getenv("LOCAL_DB_PORT"), os.Getenv("LOCAL_DB_NAME"))
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		log.Println(err)
		fmt.Println("Could not connect to database. Closing...")
		return
	}

	defer conn.Close(context.Background())
	fmt.Println("Successfully connected to database")
	r := routing.CreateRouter(conn, connectAws(), os.Getenv("BUCKET_NAME"))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://52.59.228.167:8081", "http://127.0.0.1:8081"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	})

	handler := c.Handler(r)
	fmt.Printf("Starting server at localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func main() {
	runServer(":8080")
}
