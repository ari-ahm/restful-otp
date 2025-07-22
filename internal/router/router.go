package router

import (
	"github.com/ari-ahm/restful-otp/internal/handlers"
	"github.com/gorilla/mux"
)

func NewRouter(authHandler *handlers.AuthHandler) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/initiate", authHandler.InitiateAuthHandler).Methods("POST")
	authRouter.HandleFunc("/verify", authHandler.VerifyAuthHandler).Methods("POST")

	return r
}