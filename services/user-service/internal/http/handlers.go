package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	"github.com/alexey-y-a/go-userauth-microservices/services/user-service/internal/storage"
	"github.com/rs/zerolog"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserHandler struct {
	log   zerolog.Logger
	store storage.UserStore
}

func NewUserHandler(store storage.UserStore) *UserHandler {
	log := logger.L().With().Str("service", "user-service").Str("component", "http").Logger()

	return &UserHandler{
		log:   log,
		store: store,
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users/me", h.handleGetMe)
}

func (h *UserHandler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	user, exist, err := h.store.GetUserByID(ctx, id)
	if err != nil {
		h.log.Error().Err(err).Int64("user_id", id).Msg("failed to get user by id")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	resp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.log.Error().Err(err).Int64("user_id", id).Msg("failed to write user response")
		return
	}

	h.log.Info().Int64("user_id", user.ID).Str("username", user.Username).Msg("returned user profile")
}
