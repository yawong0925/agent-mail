// sysmgr/auth.go
package sysmgr

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Verify2FARequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

func generate2FACode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// handleAdminLogin verifies the master admin's password and issues a 2FA code.
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

	// Generate 2FA code valid for 5 minutes
	code := generate2FACode()
	expiresAt := time.Now().Add(5 * time.Minute)

	_, err = s.db.Mgmt.ExecContext(r.Context(), "UPDATE local_users SET two_factor_code = ?, two_factor_expires = ? WHERE id = ?", code, expiresAt, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate 2FA"}`, http.StatusInternalServerError)
		return
	}

	// Output to console for testing purposes
	log.Printf("[SYSMGR-AUTH] Generated 2FA Code for Master Admin '%s': %s", req.Username, code)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "pending_2fa", "message": "2FA code generated. Please verify."}`))
}

// handleVerifyAdmin2FA checks the 2FA code and issues a secure session token into RAM.
func (s *Server) handleVerifyAdmin2FA(w http.ResponseWriter, r *http.Request) {
	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(req.Code)

	var userID int
	var storedCode sql.NullString
	var expiresAt sql.NullTime

	// Enforce ID 1 check again to be absolutely certain
	err := s.db.Mgmt.QueryRowContext(r.Context(), "SELECT id, two_factor_code, two_factor_expires FROM local_users WHERE username = ? AND id = 1", req.Username).Scan(&userID, &storedCode, &expiresAt)
	if err != nil || !storedCode.Valid {
		http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
		return
	}

	if storedCode.String != req.Code || time.Now().After(expiresAt.Time) {
		http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
		return
	}

	// Wipe the code from the database immediately to prevent reuse
	s.db.Mgmt.ExecContext(r.Context(), "UPDATE local_users SET two_factor_code = NULL, two_factor_expires = NULL WHERE id = ?", userID)

	sessionToken := generateSessionToken()
	sessionExpires := time.Now().Add(24 * time.Hour)

	// Save to the SysMgr's isolated RAM store
	s.sessionStore.Set(sessionToken, userID, sessionExpires)

	response := map[string]interface{}{
		"status": "success",
		"token":  sessionToken,
		"expires": sessionExpires,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}