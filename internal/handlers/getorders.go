package handlers

import (
	"encoding/json"
	"github.com/lekan/gophermart/internal/repo"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
)

func GetOrders(w http.ResponseWriter, r *http.Request) {
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

	if login == "" {
		log.Info().Msg("get orders unauthorized")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	res, err := repo.GetOrders(ctx, login)
	if err != nil {
		if len(res) == 0 {
			log.Info().Msg("no data for response")
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Err(err).Msg("get orders database error")
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(&res); err != nil {
		log.Err(err).Msg("json encoding error")
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
