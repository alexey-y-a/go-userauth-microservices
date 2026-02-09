package middleware

import (
	"context"
	"net/http"
	"strings"

	appjwt "github.com/alexey-y-a/go-userauth-microservices/libs/jwt"
	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
)


type contextKey string

const userIDContextKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
    val := ctx.Value(userIDContextKey)
    if val == nil {
        return "", false
    }
    userID, ok := val.(string)
    if !ok {
        return "", false
    }
    return userID, true
}

func JWTAuth(next http.Handler) http.Handler {
    log := logger.L().With().
        Str("service", "gateway-service").
        Str("component", "middleware").
        Logger()

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "missing Authorizathion header", http.StatusUnauthorized)
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            http.Error(w, "invalid Authorizathion header format", http.StatusUnauthorized)
            return
        }

        tokenString := parts[1]

        claims, err :=  appjwt.ParseToken(tokenString)
        if err != nil {
            log.Error().Err(err).Msg("failed to parse JWT token")
            http.Error(w, "invalid or expired token", http.StatusUnauthorized)
            return
        }

        userID := claims.UserID

        log.Info().Str("user_id", userID).Msg("JWT token validated")

        ctxWithUser := WithUserID(r.Context(), userID)

        r = r.WithContext(ctxWithUser)

        next.ServeHTTP(w, r)
    })
}