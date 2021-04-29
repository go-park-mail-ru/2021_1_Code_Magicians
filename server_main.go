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
	"github.com/jackc/pgx/v4/pgxpool"
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
				"", // a token will be created when the session is used.
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

	dbPrefix := os.Getenv("DB_PREFIX")
	if dbPrefix != "AMAZON" && dbPrefix != "LOCAL" {
		log.Fatalf("Wrong prefix: %s , should be AMAZON or LOCAL", dbPrefix)
	}

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv(dbPrefix+"_DB_USER"), os.Getenv(dbPrefix+"_DB_PASSWORD"), os.Getenv(dbPrefix+"_DB_HOST"),
		os.Getenv(dbPrefix+"_DB_PORT"), os.Getenv(dbPrefix+"_DB_NAME"))
	conn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		log.Println(err)
		fmt.Println("Could not connect to database. Closing...")
		return
	}

	defer conn.Close()
	fmt.Println("Successfully connected to database")
	r := routing.CreateRouter(conn, connectAws(), os.Getenv("BUCKET_NAME"), os.Getenv("CSRF_ON") == "true")

	allowedOrigins := make([]string, 3) // If needed, replace 3 with number of needed origins
	switch os.Getenv("HTTPS_ON") {
	case "true":
		allowedOrigins = append(allowedOrigins, "https://www.pinter-best.com:8081", "https://www.pinter-best.com", "https://127.0.0.1:8081")
	case "false":
		allowedOrigins = append(allowedOrigins, "http://www.pinter-best.com:8081", "http://www.pinter-best.com", "http://127.0.0.1:8081")
	default:
		log.Fatal("HTTPS_ON variable is not set")
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
		log.Fatal(http.ListenAndServeTLS(addr, "cert.pem", "key.pem", handler))
	case "false":
		log.Fatal(http.ListenAndServe(addr, handler))
	}
}

func main() {
	runServer(":8080")
}
