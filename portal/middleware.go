// portal/middleware.go
package portal

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type contextKey string
const userIDKey contextKey = "user_id"

// RequireAuth intercepts API calls and validates the RAM session token.
func (s *Server) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Unauthorized: Missing token"}`, http.StatusUnauthorized)
			return
		}

		// 2. Extract the raw token string
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Perform a nanosecond lookup directly from RAM
		session, exists := s.sessionStore.Get(token)
		
		if !exists {
			http.Error(w, `{"error": "Unauthorized: Invalid session"}`, http.StatusUnauthorized)
			return
		}

		// 4. Verify expiration
		if time.Now().After(session.ExpiresAt) {
			// Purge the expired token from RAM immediately
			s.sessionStore.Delete(token)
			http.Error(w, `{"error": "Unauthorized: Session expired"}`, http.StatusUnauthorized)
			return
		}

		// 5. Inject the UserID into the HTTP context so the route handler can use it
		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}