package mware

import (
	"github.com/lekan/gophermart/internal/logger"
	"github.com/lekan/gophermart/internal/sessions"
	"net/http"
)

var log = logger.GetLogger()

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
			//http.Error(w, err.Error(), 500)
			w.WriteHeader(http.StatusInternalServerError)
			//next.ServeHTTP(w, r)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			log.Info().Msg("access denied")
			//http.Error(w, err.Error(), http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			//next.ServeHTTP(w, r)
			return
		}
		if login, ok := session.Values["login"].(string); !ok || login == "" {
			log.Info().Msg("unknown login")
			//http.Error(w, err.Error(), http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			//next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
