// internal/http/middleware/recovery.go
package middleware

import (
	"net/http"

	"reviewer-service/internal/http/response"

	"github.com/rs/zerolog/log"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Msg("panic recovered")

				response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
