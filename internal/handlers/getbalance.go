package handlers

import (
	"context"
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	session, err := sessions.Get(r)
	if err != nil {
		log.Err(err).Msg("Session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	value := session.Values["login"]
	login, ok := value.(string)
	if !ok {
		log.Info().Msg("type assertion error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balance, err := models.GetBalance(ctx, login)
	if err != nil {
		log.Err(err).Msg("get balance error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&balance); err != nil {
		log.Err(err).Msg("json encoding error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
