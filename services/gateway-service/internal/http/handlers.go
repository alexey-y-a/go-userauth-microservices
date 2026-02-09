package http

import (
	"net/http"

	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	"github.com/alexey-y-a/go-userauth-microservices/services/gateway-service/internal/middleware"
	"github.com/rs/zerolog"
)

type GatewayHandler struct {
    log zerolog.Logger
}

func NewGatewayHandler() *GatewayHandler {
    log := logger.L().With().
        Str("service", "gateway-service").
        Str("component", "http").
        Logger()

    return &GatewayHandler{
        log: log,
    }
}

func (h *GatewayHandler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/public", h.handlePublic)

    protectedHandler := http.HandlerFunc(h.handleProtected)
    protectedWithAuth := middleware.JWTAuth(protectedHandler)
    mux.Handle("/protected", protectedWithAuth)
}

func (h *GatewayHandler) handlePublic(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, err := w.Write([]byte("public endpoint, no auth required\n"))
    if err != nil {
        h.log.Error().Err(err).Msg("failed to write public response")
        return
    }
}

func (h *GatewayHandler) handleProtected(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "user id not found in context", http.StatusUnauthorized)
        return
    }

    h.log.Info().Str("user_id", userID).Msg("access to protected endpoint")

    w.WriteHeader(http.StatusOK)
    _, err := w.Write([]byte("hello from protected endpoint, user_id=" + userID + "\n"))
    if err != nil {
        h.log.Error().Err(err).Msg("failed to write protected response")
        return
    }

}