package mware

import (
	"context"
	"net/http"
	"time"
)

func SetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		go func() {
			select {
			case <-time.After(5 * time.Second):
				cancel()
				return
			case <-ctx.Done():
				return
			}
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
