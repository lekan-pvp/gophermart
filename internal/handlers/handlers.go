package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/lekan/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("token_password")))

func Signup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	creds := &models.Credentials{}

	// получаем Body
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		log.Println("JSON error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		log.Println("Hashing error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// обновляем пароль в аккаунте
	creds.Password = string(hashedPassword)

	// сохраняем в базу данных
	err = models.Signup(ctx, creds)
	if err != nil {
		if errors.Is(err, fmt.Errorf("409 %w", err)) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("Signup error")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// создаем сессию
	session, err := store.Get(r, "session_token")
	if err != nil {
		log.Println("Session error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// записываем токен в сессию и сохраняем сессию
	session.Values[creds.Login] = hashedPassword
	err = session.Save(r, w)
	if err != nil {
		log.Println("Save session error")
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

	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := models.Signin(ctx, creds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

}
