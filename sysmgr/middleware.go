// sysmgr/middleware.go
package sysmgr

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type contextKey string
const userIDKey contextKey = "user_id"

// RequireAuth protects sysmgr endpoints by verifying the RAM-based session token.
func (s *Server) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Unauthorized: Missing token"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Look up the token in the SysMgr's specific RAM map
		session, exists := s.sessionStore.Get(token)
		
		if !exists {
			http.Error(w, `{"error": "Unauthorized: Invalid session"}`, http.StatusUnauthorized)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			s.sessionStore.Delete(token)
			http.Error(w, `{"error": "Unauthorized: Session expired"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Middleware for setting CORS headers if you run the Vue dev server on a different port (like 5173).
func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}