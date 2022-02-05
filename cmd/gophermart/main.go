package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/mware"
	"log"
	"net/http"
)

func main() {
	c := cfg.GetConfig()

	err := models.InitDB(c.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(mware.JWTAuthentication)
	router.Use(middleware.Logger)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.CreateAccount)
	})

	log.Println("running server...")
	err = http.ListenAndServe(c.RunAddress, router)
	if err != nil {
		fmt.Println(err)
	}
}
