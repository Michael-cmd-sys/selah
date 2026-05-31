package middleware

import (
	"net/http"
	"strings"

	"github.com/selah/internal/auth"
	"github.com/selah/internal/response"
)

// Authenticate is a strict middleware — rejects unauthenticated requests.
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := extractClaims(r, jwtSecret)
			if err != nil {
				response.Unauthorized(w, "authentication required")
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), claims.UserID)))
		})
	}
}

// OptionalAuth attaches user identity if a valid token is present, but never blocks.
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if claims, err := extractClaims(r, jwtSecret); err == nil {
				r = r.WithContext(auth.WithUserID(r.Context(), claims.UserID))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractClaims(r *http.Request, secret string) (*auth.Claims, error) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return nil, http.ErrNoCookie
	}
	return auth.ParseToken(token, secret)
}
