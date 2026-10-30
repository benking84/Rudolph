package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/airbnb/rudolph/internal/handlers"
)

func main() {
	if err := handlers.Boot(); err != nil {
		log.Fatalf("boot: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handlers.HTTPHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("rudolph listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
