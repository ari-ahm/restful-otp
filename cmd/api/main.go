package main

import (
	"log"
	"net/http"
	"github.com/ari-ahm/restful-otp/internal/router"
)

func main() {
	r := router.NewRouter()

	port := ":8080"

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}