package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lekan/gophermart/internal/cfg"
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/lekan/gophermart/internal/logger"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/mware"
	"net/http"
)

func main() {
	c := cfg.GetConfig()
	log := logger.GetLogger()

	err := models.InitDB(c.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err)
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(mware.CheckUser)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.Signup)
		r.Post("/login", handlers.Signin)
		r.Get("/balance", handlers.GetBalance)
		r.Get("/withdrawals", handlers.GetWithdrawals)
	})

	log.Info().Msg("server is up...")
	err = http.ListenAndServe(c.RunAddress, router)
	if err != nil {
		log.Fatal().Err(err)
	}
}
