package handlers

import (
	"encoding/json"
	"github.com/lekan/gophermart/internal/repo"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
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

	req := &repo.Wdraw{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Err(err).Msg("json decode error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode, err := repo.Withdraw(ctx, login, req)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}
