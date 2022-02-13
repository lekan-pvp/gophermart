package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	creds := &models.Credentials{}

	// получаем Body
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		log.Err(err).Msg("JSON error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// сохраняем в базу данных
	err := models.Signup(ctx, creds)
	if err != nil {
		if errors.Is(err, fmt.Errorf("409 %w", err)) {
			log.Info().Msg("Login is in use another user")
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Err(err).Msg("Signup error")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// создаем сессию
	session, err := sessions.Get(r)
	if err != nil {
		log.Err(err).Msg("Session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// записываем токен в сессию и сохраняем сессию
	session.Values["authenticated"] = true
	session.Values["login"] = creds.Login
	err = session.Save(r, w)
	if err != nil {
		log.Err(err).Msg("Save session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Context-Type", "application/json")
	w.WriteHeader(200)
}
