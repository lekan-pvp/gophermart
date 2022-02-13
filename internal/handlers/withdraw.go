package handlers

import (
	"context"
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
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

	req := &models.Wdraw{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Err(err).Msg("json decode error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode, err := models.Withdraw(ctx, login, req)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}
