package main

import (
	"fmt"
	"log"
	"net/http"
	"pinterest/routing"

	"github.com/rs/cors"
)

func runServer(addr string) {
	r := routing.CreateRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://52.59.228.167:8081"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	fmt.Printf("Starting server at localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func main() {
	runServer(":8080")
}
