package mware

import (
	"context"
	"net/http"
)

func SetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		go func() {
			select {
			case <-ctx.Done():
				return
			}
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
