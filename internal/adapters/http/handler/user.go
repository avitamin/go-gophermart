// Package handler provides HTTP handlers for the application.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/avitamin/go-gophermart/internal/adapters/http/middleware"
	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"go.uber.org/zap"
)

// DefaultTokenDuration is the default token duration for cookies.
const DefaultTokenDuration = 7 * 24 * time.Hour

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *service.UserService
	log         *zap.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// setTokenCookie sets the JWT token in a cookie.
func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.TokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(DefaultTokenDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Register handles user registration.
// POST /api/user/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds entity.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Register(r.Context(), &creds)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		if errors.Is(err, entity.ErrPasswordTooShort) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.Error("registration failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// Login handles user authentication.
// POST /api/user/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds entity.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), &creds)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.log.Error("login failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
