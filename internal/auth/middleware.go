package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"confluent/internal/models"
)

type contextKey string

const userContextKey = contextKey("auth_user")

// Middleware creates HTTP middleware to extract and authenticate JWT tokens.
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := ParseToken(tokenStr, secret)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user := &models.User{
				ID:       claims.UserID,
				Email:    claims.Email,
				FullName: claims.FullName,
				Role:     claims.Role,
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ContextWithUser injects the authenticated user into a context (useful for testing and internal calls).
func ContextWithUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext returns the authenticated user if present in the context.
func UserFromContext(ctx context.Context) *models.User {
	u, ok := ctx.Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return u
}

// RequireAuth ensures that the incoming request is authenticated.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil {
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		next(w, r)
	}
}

// RequireRole ensures that the user has at least one of the permitted roles.
func RequireRole(roles ...models.Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u := UserFromContext(r.Context())
			if u == nil {
				writeJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			hasRole := false
			for _, r := range roles {
				if u.Role == r {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeJSONError(w, http.StatusForbidden, "You do not have permission to perform this action")
				return
			}

			next(w, r)
		}
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	// Fallback to cookie
	if c, err := r.Cookie("token"); err == nil {
		return c.Value
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
