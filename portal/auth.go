// portal/auth.go
package portal

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

	// bcrypt is the industry standard for securely hashing and verifying passwords.
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest defines the expected JSON payload for step 1 of authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Verify2FARequest defines the expected JSON payload for step 2 of authentication.
type Verify2FARequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

// generate2FACode creates a random 6-character alphanumeric string.
// We exclude ambiguous characters (like 0 and O, 1 and I) if desired, but here we use a standard A-Z 0-9 set.
func generate2FACode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// generateSessionToken creates a highly secure 64-character hexadecimal session string.
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// handleLogin validates the user's password and issues a 2FA code to the database.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// 1. Fetch the user's ID and hashed password from DB3 (Mgmt)
	var userID int
	var hash string
	err := s.db.Mgmt.QueryRowContext(r.Context(), "SELECT id, password_hash FROM local_users WHERE username = ?", req.Username).Scan(&userID, &hash)
	
	if err == sql.ErrNoRows {
		// Generic error prevents attackers from knowing which usernames exist in the system.
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 2. Compare the provided plain-text password against the stored bcrypt hash.
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// 3. Generate the 6-character 2FA code and set expiration to exactly 5 minutes from now.
	code := generate2FACode()
	expiresAt := time.Now().Add(5 * time.Minute)

	// Update the user's record in the database with the new 2FA requirements.
	_, err = s.db.Mgmt.ExecContext(r.Context(), "UPDATE local_users SET two_factor_code = ?, two_factor_expires = ? WHERE id = ?", code, expiresAt, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate 2FA session"}`, http.StatusInternalServerError)
		return
	}

	// 4. Send the code. In production, this would trigger an SMTP email or SMS API.
	// For now, we log it to the console so you can test the login flow.
	log.Printf("[AUTH] Generated 2FA Code for %s: %s (Expires in 5 minutes)", req.Username, code)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "pending_2fa", "message": "2FA code sent to your email. Please verify."}`))
}

// handleVerify2FA checks the 6-digit code and issues the in-memory session token.
func (s *Server) handleVerify2FA(w http.ResponseWriter, r *http.Request) {
	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Force the incoming code to uppercase so the check is case-insensitive.
	req.Code = strings.ToUpper(req.Code)

	// 1. Fetch the user's current 2FA state from the database.
	var userID int
	var storedCode sql.NullString // Use NullString in case the code is currently NULL
	var expiresAt sql.NullTime

	err := s.db.Mgmt.QueryRowContext(r.Context(), "SELECT id, two_factor_code, two_factor_expires FROM local_users WHERE username = ?", req.Username).Scan(&userID, &storedCode, &expiresAt)
	if err != nil || !storedCode.Valid {
		http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
		return
	}

	// 2. Validate the code matches and that the 5-minute window hasn't expired.
	if storedCode.String != req.Code || time.Now().After(expiresAt.Time) {
		http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
		return
	}

	// 3. Clear the 2FA code from the database immediately to prevent replay attacks.
	s.db.Mgmt.ExecContext(r.Context(), "UPDATE local_users SET two_factor_code = NULL, two_factor_expires = NULL WHERE id = ?", userID)

	// 4. Generate a secure session token valid for 24 hours.
	sessionToken := generateSessionToken()
	sessionExpires := time.Now().Add(24 * time.Hour)

	// 5. Save the session token to the blazing-fast MemorySessionStore (RAM) instead of SQLite.
	s.sessionStore.Set(sessionToken, userID, sessionExpires)

	// 6. Return the session token to the frontend. The frontend should attach this as a 'Bearer' token in subsequent requests.
	response := map[string]interface{}{
		"status": "success",
		"token":  sessionToken,
		"expires": sessionExpires,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}