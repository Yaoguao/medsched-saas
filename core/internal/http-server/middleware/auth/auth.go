package authmiddleware

import (
	"base/internal/lib/jwt"
	"context"
	"net/http"
	"strings"
)

func MiddlewareAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokenStr := extractToken(r)

			claims, err := jwt.ParseToken(tokenStr, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "uid", claims.UserID.String())
			ctx = context.WithValue(ctx, "role", string(claims.Role))
			ctx = context.WithValue(ctx, "email", claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {

	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")

	if len(parts) != 2 {
		return ""
	}

	if parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
