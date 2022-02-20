package handlers

import (
	"encoding/json"
	"github.com/lekan/gophermart/internal/repo"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	balance, err := repo.GetBalance(ctx, login)
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
