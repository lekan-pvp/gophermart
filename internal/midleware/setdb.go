package midleware

import (
	"context"
	"github.com/lekan/gophermart/internal/cfg"
	"net/http"
	"time"
)

func SetDBMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeoutContext, _ := context.WithTimeout(context.Background(), time.Second)
		ctx := context.WithValue(r.Context(), "DB", cfg.GetDB().WithContext(timeoutContext))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
