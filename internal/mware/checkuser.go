package mware

import (
	"github.com/lekan/gophermart/internal/sessions"
	"github.com/rs/zerolog/log"
	"net/http"
)

func CheckUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notAuth := []string{"/api/user/register", "/api/user/login"}
		requestPath := r.URL.Path
		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		session, err := sessions.Get(r)
		if err != nil {
			log.Err(err).Msg("session initialization error")
			http.Error(w, err.Error(), 500)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			log.Error().Msg("access denied")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if login, ok := session.Values["login"].(string); !ok || login == "" {
			log.Err(err).Msg("unknown login")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}