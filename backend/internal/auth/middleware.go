package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
)

type contextKey struct{}

// RequireAuth rejects requests without a valid "Authorization: Bearer <token>" header
// and makes the token's claims available via ClaimsFromContext
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || strings.TrimSpace(tokenStr) == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims, err := Authenticate(r.Context(), strings.TrimSpace(tokenStr))
		if err != nil {
			if !errors.Is(err, ErrInvalidToken) {
				log.Printf("authenticate: %v", err)
			}
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims)))
	}
}

// ClaimsFromContext returns the claims set by RequireAuth
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(contextKey{}).(*Claims)
	return claims
}
