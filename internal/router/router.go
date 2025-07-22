package router

import (
	"github.com/ari-ahm/restful-otp/internal/handlers"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")

	return r
}