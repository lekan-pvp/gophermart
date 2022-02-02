package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/lekan/gophermart/internal/midleware"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()
	router.Use(midleware.JWTAuthentication)

	router.Route("api/user", func(r chi.Router) {
		r.Post("/register", handlers.CreateAccount)
	})

	c := cfg.GetConfig()

	log.Println("running server...")
	err := http.ListenAndServe(c.RunAddress, router)
	if err != nil {
		fmt.Println(err)
	}
}
