package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	appjwt "github.com/alexey-y-a/go-userauth-microservices/libs/jwt"
	"github.com/alexey-y-a/go-userauth-microservices/libs/logger"
	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password"`
}

func (r RegisterRequest) Validate() error {
    return validation.ValidateStruct(
        &r,
        validation.Field(&r.Username, validation.Required, validation.Length(3, 50)),
        validation.Field(&r.Email, validation.Required, is.Email),
        validation.Field(&r.Password, validation.Required, validation.Length(8, 100)),
    )
}

type RegisterResponse struct {
    ID int64 `json:"id"`
    Username string `json:"username"`
    Email string `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type AuthHandler struct {
    log zerolog.Logger
    store storage.Store
}

func NewAuthHandler(store storage.Store) *AuthHandler {
    log := logger.L().With().
        Str("service", "auth-service").
        Str("component", "http").
        Logger()

    return  &AuthHandler{
        log: log,
        store: store,
    }
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/auth/register", h.handleRegister)
    mux.HandleFunc("/auth/login", h.handleLogin)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.Header().Set("Allow", http.MethodPost)
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req RegisterRequest

    err := json.NewDecoder(r.Body).Decode(&req)
    defer r.Body.Close()

    if err != nil {
        h.log.Error().Err(err).Msg("failed to decode register request")
        http.Error(w, "invalid JSON body", http.StatusBadRequest)
        return
    }

    err = req.Validate()
    if err != nil {
        h.log.Error().Err(err).Msg("validation failed for register request")
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        h.log.Error().Err(err).Msg("failed to hash password")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    passwordHash := string(passwordHashBytes)

    user, err := h.store.CreateUser(req.Username, req.Email, passwordHash)
    if err != nil {
        if err == storage.ErrUserAlreadyExists {
            h.log.Error().Err(err).Str("username", req.Username).Msg("user already exists")
            http.Error(w, "user already exists", http.StatusConflict)
            return
        }

        h.log.Error().Err(err).Msg("failed to create user")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    resp := RegisterResponse {
        ID: user.ID,
        Username: user.Username,
        Email: user.Email,
        CreatedAt: time.Now().UTC(),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    err = json.NewEncoder(w).Encode(resp)
    if err != nil {
        h.log.Error().Err(err).Msg("failed to write register response")
        return
    }

    h.log.Info().
        Str("username", user.Username).
        Int64("user_id", user.ID).
        Msg("user registered")
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
    return validation.ValidateStruct(
        &r,
        validation.Field(&r.Username, validation.Required, validation.Length(3, 50)),
        validation.Field(&r.Password, validation.Required, validation.Length(8, 100)),
    )
}

type LoginResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresAt time.Time `json:"expires_at"`
    ID int64 `json:"id"`
    Username string `json:"username"`
    Email string `json:"email"`
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.Header().Set("Allow", http.MethodPost)
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req LoginRequest

    err := json.NewDecoder(r.Body).Decode(&req)
    defer r.Body.Close()

    if err != nil {
        h.log.Error().Err(err).Msg("failed to decode login request")
        http.Error(w, "invalid JSON body", http.StatusBadRequest)
        return
    }

    err = req.Validate()
    if err != nil {
        h.log.Error().Err(err).Msg("validation failed for login request")
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    user, exist , err := h.store.GetUserByUsername(req.Username)
    if err != nil {
        h.log.Error().Err(err).Msg("failed to get user by username")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    if !exist {
        h.log.Warn().Str("username", req.Username).Msg("user not found on login")
        http.Error(w, "invalid username or password", http.StatusUnauthorized)
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    if err != nil {
        h.log.Warn().Str("username", req.Username).Msg("invalid password or login")
        http.Error(w, "invalid username or password", http.StatusUnauthorized)
        return
    }

    accessToken, err := appjwt.GenerateAccessToken(userIDToString(user.ID))
    if err != nil {
        h.log.Error().Err(err).Int64("user_id", user.ID).Msg("failed to generate access token")
        http.Error(w, "failed to generate token", http.StatusInternalServerError)
        return
    }

    now := time.Now().UTC()

    resp := LoginResponse {
        AccessToken: accessToken,
        ExpiresAt: now.Add(time.Hour),
        ID: user.ID,
        Username: user.Username,
        Email: user.Email,
    }

    w.Header().Set("Content-Type", "application-json")
    w.WriteHeader(http.StatusOK)

    err = json.NewEncoder(w).Encode(resp)
    if err != nil {
        h.log.Error().Err(err).Msg("failed to write login response")
        return
    }

    h.log.Info().
        Int64("user_id", user.ID).
        Str("username", user.Username).
        Msg("user logged in")
}

func userIDToString(id int64) string {
    return strconv.FormatInt(id, 10)
}