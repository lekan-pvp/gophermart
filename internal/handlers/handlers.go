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
	"io"
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
	defer r.Body.Close()

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

func Orders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	orderId, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Err(err).Msg("take orderId error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	statusCode, err := models.PostOrder(ctx, login, orderId)
	if err != nil {
		log.Err(err).Int("PostOrder error, status code: %d", statusCode)
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(statusCode)
}

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

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		log.Err(err).Msg("json encoding error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	res, err := models.GetWithdrawals(ctx, login)
	if err != nil {
		log.Err(err).Msg("database error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) == 0 {
		log.Info().Msg("nothing withdraw")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(&res); err != nil {
		log.Err(err).Msg("json encoder error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

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
		log.Err(err).Msg("database error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) == 0 {
		log.Info().Msg("no data for response")
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

//func Withdraw(w http.ResponseWriter, r *http.Request) {
//	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
//	defer cancel()
//
//	session, err := sessions.Get(r)
//	if err != nil {
//		log.Err(err).Msg("Session error")
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//}
