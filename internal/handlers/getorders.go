package handlers

import (
	"context"
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

func GetOrders(w http.ResponseWriter, r *http.Request) {
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

	res, err := models.GetOrders(ctx, login)
	if err != nil {
		log.Err(err).Msg("get orders database error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) == 0 {
		log.Info().Msg("no data for response")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(res); err != nil {
		log.Err(err).Msg("json encoding error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
