package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			token, ok := extractBearerToken(authHeader)

			if !ok {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid token")
				return
			}

			claims, err := parseToken(token, secret)

			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}