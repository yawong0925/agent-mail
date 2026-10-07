// sysmgr/auth.go
package sysmgr

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// handleAdminLogin verifies the master admin's password and issues a session token.
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	var userID int
	var hash string
	
	// SECURITY: We explicitly check if this user is ID 1 (the master admin)
	err := s.db.Mgmt.QueryRowContext(r.Context(), "SELECT id, password_hash FROM local_users WHERE username = ? AND id = 1", req.Username).Scan(&userID, &hash)
	
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid credentials or not a master admin"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	sessionToken := generateSessionToken()
	sessionExpires := time.Now().Add(24 * time.Hour)

	// Save to the SysMgr's isolated RAM store
	s.sessionStore.Set(sessionToken, userID, sessionExpires)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"token":   sessionToken,
		"expires": sessionExpires,
	})
}
