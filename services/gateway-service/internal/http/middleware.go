package http

import (
	"context"
	"net/http"
	"strings"

	appjwt "github.com/alexey-y-a/go-userauth-microservices/libs/jwt"
)

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		claims, err := appjwt.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		userId := claims.Subject

		if userId == "" {
			http.Error(w, "invalid token: missing subject", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userId)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
