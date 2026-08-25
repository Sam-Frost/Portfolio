package auth

import (
	"net/http"
	"strings"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/httpx"
)

// RequireAuth wraps next so it only runs when the request carries a valid
// "Authorization: Bearer <token>" header, as issued by Service.Login.
func RequireAuth(service *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				httpx.WriteError(w, apperr.Unauthorized("missing bearer token"))
				return
			}

			if err := service.Verify(token); err != nil {
				httpx.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
