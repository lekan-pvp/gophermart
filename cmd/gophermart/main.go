package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lekan/gophermart/internal/config"
	"github.com/lekan/gophermart/internal/handlers"
	"github.com/lekan/gophermart/internal/logger"
	"github.com/lekan/gophermart/internal/mware"
	"github.com/lekan/gophermart/internal/repo"
	"github.com/rs/zerolog"
	"net/http"
)

var log zerolog.Logger

var c *config.Config

func main() {
	logger.InitLogger()
	c = config.New()
	log = logger.New()
	err := repo.New(c.DatabaseURI)
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
		r.Get("/orders", handlers.GetOrders)
		r.Post("/orders", handlers.Orders)
		r.Post("/balance/withdraw", handlers.Withdraw)
	})

	log.Info().Msg("server is up...")
	err = http.ListenAndServe(c.RunAddress, router)
	if err != nil {
		log.Fatal().Err(err)
	}
}
