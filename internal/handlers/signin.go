package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/lekan/gophermart/internal/repo"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

func Signin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	creds := &repo.Credentials{}

	//получаем body
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		log.Err(err).Msg("json error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// ищем в базе данных
	err := repo.Signin(ctx, creds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg("user does not exist")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := sessions.Get(r)
	if err != nil {
		log.Err(err).Msg("Session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = true
	session.Values["login"] = creds.Login
	err = session.Save(r, w)
	if err != nil {
		log.Err(err).Msg("save session error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Context-Type", "application/json")
	w.WriteHeader(200)
}
