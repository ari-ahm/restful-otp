package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ari-ahm/restful-otp/internal/handlers"
	"github.com/ari-ahm/restful-otp/internal/repository"
	"github.com/ari-ahm/restful-otp/internal/router"
	"github.com/ari-ahm/restful-otp/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	port := os.Getenv("PORT")

	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET_KEY environment variable not set")
	}
	if port == "" {
		port = "8080"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	log.Println("Successfully connected to the database.")

	userRepo := repository.NewUserRepository(dbpool)
	authService := services.NewAuthService(userRepo, jwtSecret)
	authHandler := handlers.NewAuthHandler(authService)

	r := router.NewRouter(authHandler)
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}