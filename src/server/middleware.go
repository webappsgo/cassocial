package server

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyUsername is the context key for username
	ContextKeyUsername ContextKey = "username"
	// ContextKeyRole is the context key for user role
	ContextKeyRole ContextKey = "role"
	// ContextKeyClaims is the context key for full JWT claims
	ContextKeyClaims ContextKey = "claims"
)

// Middleware provides authentication middleware functionality
type Middleware struct {
	auth *Auth
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware(auth *Auth) *Middleware {
	return &Middleware{auth: auth}
}

// RequireAuth is middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := m.extractToken(r)
		if token == "" {
			m.unauthorizedResponse(w, "missing authorization token")
			return
		}

		// Validate token
		claims, err := m.auth.ValidateToken(token)
		if err != nil {
			m.unauthorizedResponse(w, "invalid or expired token")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is middleware that requires a specific role
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, ensure user is authenticated
			userRole, ok := r.Context().Value(ContextKeyRole).(string)
			if !ok {
				m.unauthorizedResponse(w, "authentication required")
				return
			}

			// Check if user has required role
			if userRole != role {
				m.forbiddenResponse(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is middleware that requires admin role
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole("admin")(next)
}

// RequireActiveUser is middleware that ensures the user is active
func (m *Middleware) RequireActiveUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, ok := r.Context().Value(ContextKeyUserID).(string)
		if !ok {
			m.unauthorizedResponse(w, "authentication required")
			return
		}

		// Get user from database
		user, err := m.auth.GetUserByID(userID)
		if err != nil {
			m.unauthorizedResponse(w, "user not found")
			return
		}

		// Check if user is active
		if !user.IsActive() {
			m.forbiddenResponse(w, "user account is not active")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuth is middleware that optionally authenticates if token is present
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := m.extractToken(r)
		if token == "" {
			// No token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		claims, err := m.auth.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitByUser is middleware that applies rate limiting per user
func (m *Middleware) RateLimitByUser(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (may be empty for unauthenticated requests)
			userID, _ := r.Context().Value(ContextKeyUserID).(string)
			if userID == "" {
				// Use IP address for unauthenticated users
				userID = getIPAddress(r)
			}

			// Check rate limit
			if !limiter.Allow(userID) {
				m.rateLimitResponse(w, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORS is middleware that handles CORS headers
func (m *Middleware) CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders is middleware that adds security headers
func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers as per SPEC
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")

		next.ServeHTTP(w, r)
	})
}

// ExtractToken extracts the JWT token from the Authorization header or cookie.
// Exported so that the handler sub-package can use it for optional auth checks.
func (m *Middleware) ExtractToken(r *http.Request) string {
	return m.extractToken(r)
}

// extractToken extracts the JWT token from the Authorization header
func (m *Middleware) extractToken(r *http.Request) string {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try to get from cookie as fallback
		cookie, err := r.Cookie("token")
		if err != nil {
			return ""
		}
		return cookie.Value
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// unauthorizedResponse sends a 401 Unauthorized response
func (m *Middleware) unauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// forbiddenResponse sends a 403 Forbidden response
func (m *Middleware) forbiddenResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// rateLimitResponse sends a 429 Too Many Requests response
func (m *Middleware) rateLimitResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// GetUserIDFromContext extracts the user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(string)
	return userID, ok
}

// GetUsernameFromContext extracts the username from request context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(ContextKeyUsername).(string)
	return username, ok
}

// GetRoleFromContext extracts the role from request context
func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ContextKeyRole).(string)
	return role, ok
}

// GetClaimsFromContext extracts the full JWT claims from request context
func GetClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*JWTClaims)
	return claims, ok
}

// IsAdmin checks if the user in context is an admin
func IsAdmin(ctx context.Context) bool {
	role, ok := GetRoleFromContext(ctx)
	return ok && role == "admin"
}
