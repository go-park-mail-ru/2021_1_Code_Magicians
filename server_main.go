package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pinterest/interfaces/routing"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func runServer(addr string) {
	godotenv.Load(".env")
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
	r := routing.CreateRouter(conn)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://52.59.228.167:8081"},
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
