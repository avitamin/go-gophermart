// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/avitamin/go-gophermart/internal/pkg/auth"
)

type contextKey string

// UserIDKey is the context key for user ID.
const UserIDKey contextKey = "userID"

// TokenCookieName is the name of the cookie that stores the JWT token.
const TokenCookieName = "token"

// Auth creates authentication middleware.
// It checks for JWT token in Authorization header first, then in cookie.
func Auth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// Try to get token from Authorization header first
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// If no token in header, try to get from cookie
			if token == "" {
				cookie, err := r.Cookie(TokenCookieName)
				if err == nil && cookie.Value != "" {
					token = cookie.Value
				}
			}

			// If still no token, return unauthorized
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.Verify(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
