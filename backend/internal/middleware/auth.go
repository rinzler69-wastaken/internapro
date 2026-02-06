package middleware

import (
	"context"
	"net/http"
	"strings"

	"dsi_interna_sys/internal/config"
	"dsi_interna_sys/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

// Claims represents JWT claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(w, "Missing authorization header")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondUnauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Loaded.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			utils.RespondUnauthorized(w, "Invalid or expired token")
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				utils.RespondUnauthorized(w, "Unauthorized")
				return
			}

			normalized := NormalizeRole(claims.Role)

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if normalized == NormalizeRole(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				utils.RespondForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NormalizeRole maps legacy role names to the canonical set.
func NormalizeRole(role string) string {
	if role == "supervisor" {
		return "pembimbing"
	}
	return role
}
