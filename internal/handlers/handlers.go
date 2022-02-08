package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lekan/gophermart/internal/logger"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
	"time"
)

var log = logger.GetLogger()

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
	session.Values["token"] = creds.Login
	err = session.Save(r, w)
	if err != nil {
		log.Err(err).Msg("Save session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Context-Type", "application/json")
	w.WriteHeader(200)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	creds := &models.Credentials{}

	//получаем body
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		log.Err(err).Msg("json error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ищем в базе данных
	err := models.Signin(ctx, creds)
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

	session.Values["token"] = creds.Login
	err = session.Save(r, w)
	if err != nil {
		log.Err(err).Msg("save session error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Context-Type", "application/json")
	w.WriteHeader(200)
}
