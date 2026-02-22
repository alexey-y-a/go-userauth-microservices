package http

import (
	"io"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-userauth-microservices/libs/httpclient"
	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	"github.com/alexey-y-a/go-userauth-microservices/services/gateway-service/internal/middleware"
	"github.com/rs/zerolog"
)

type GatewayHandler struct {
	log    zerolog.Logger
	client *httpclient.Client
}

func NewGatewayHandler() *GatewayHandler {
	log := logger.L().With().
		Str("service", "gateway-service").
		Str("component", "http").
		Logger()

	client := httpclient.NewDefaultClient()

	return &GatewayHandler{
		log:    log,
		client: client,
	}
}

func (h *GatewayHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/public", h.handlePublic)

	protectedHandler := http.HandlerFunc(h.handleProtected)
	protectedWithAuth := JWTAuthMiddleware(protectedHandler)
	mux.Handle("/protected", protectedWithAuth)

	meHandler := http.HandlerFunc(h.handleMe)
	meWithAuth := JWTAuthMiddleware(meHandler)
	mux.Handle("/me", meWithAuth)

	mux.HandleFunc("/health/user", h.handleUserServiceHealth)
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

func (h *GatewayHandler) handleUserServiceHealth(w http.ResponseWriter, r *http.Request) {
	userServiceURL := "http://localhost:8081/health"

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, userServiceURL, nil)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to create request to user-service health")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp, err := h.client.DoWithRetry(req)
	if err != nil {
		h.log.Error().Err(err).Str("url", userServiceURL).Msg("failed to call user-service health")
		http.Error(w, "user-service unavailable", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.log.Error().Int("status_code", resp.StatusCode).Str("url", userServiceURL).
			Msg("user-service health returned non-200")
		http.Error(w, "user-service unhealthy", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("user-service is healthy\n"))
	if err != nil {
		h.log.Error().Err(err).Msg("failed to write user-service health proxy response")
		return
	}
}

func (h *GatewayHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(userIDContextKey)
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		http.Error(w, "user id not found in context", http.StatusUnauthorized)
		return
	}

	userServiceUrl := "http://localhost:8081/users/me?id=" + userID

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, userServiceUrl, nil)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to create request to user-service /users/me")
		http.Error(w, "internal errpr", http.StatusInternalServerError)
		return
	}

	resp, err := h.client.DoWithRetry(req)
	if err != nil {
		h.log.Error().Err(err).Str("url", userServiceUrl).Msg("failed to call user-service /users/me")
		http.Error(w, "user-service unavailable", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.log.Warn().Int("status_code", resp.StatusCode).Str("url", userServiceUrl).Msg("user-service /users/me returned non-200")

		if resp.StatusCode == http.StatusNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(w, "user-service error", http.StatusBadGateway)
		return
	}

	for k, vs := range resp.Header {
		if strings.EqualFold(k, "Content-Type") {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}

	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to proxy /users/me response body")
		return
	}

}
