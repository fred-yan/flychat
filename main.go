package main

import (
	"flychat/api"
	"flychat/lib"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	port = ":8080"
)

func init() {
	lib.InstallDB()
}

func main() {
	fmt.Println("Server started...")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	}).Handler)

	apiRouter := chi.NewRouter()
	auth := &api.Auth{}
	apiRouter.Post("/signup", auth.Signup)
	apiRouter.Post("/signin", auth.Signin)
	router.Mount("/api", apiRouter)

	if err := http.ListenAndServe(port, router); err != nil {
		log.Printf("HTTP server failed to start on port %s: %v", port, err)
	}
}
