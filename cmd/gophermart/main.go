package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/lekan/gophermart/internal/models"
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
	router.Use(middleware.Logger)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.Signup)
		r.Post("/login", handlers.Signin)
	})

	log.Println("running server...")
	err = http.ListenAndServe(c.RunAddress, router)
	if err != nil {
		fmt.Println(err)
	}
}
