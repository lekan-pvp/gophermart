package handlers

import (
	"context"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/sessions"
	"io"
	"net/http"
)

func Orders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
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

	if login == "" {
		log.Info().Msg("post order unauthorized")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orderId, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Err(err).Msg("take orderId error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
