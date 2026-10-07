// sysmgr/setup.go
package sysmgr

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type SetupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleSetupStatus tells the frontend if it needs to display the initialization screen.
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	var count int
	// We count local_users to see if anyone exists yet
	err := s.db.Mgmt.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM local_users").Scan(&count)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"setup_required": count == 0,
	})
}

// handleSetupAdmin processes the first-time setup form.
func (s *Server) handleSetupAdmin(w http.ResponseWriter, r *http.Request) {
	var count int
	s.db.Mgmt.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM local_users").Scan(&count)
	
	// Hard block: If even 1 user exists, this endpoint shuts down permanently.
	if count > 0 {
		http.Error(w, `{"error": "System already initialized"}`, http.StatusForbidden)
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// In the future, we will add an 'is_admin' column, but for now, the first user is the admin.
	insertSQL := "INSERT INTO local_users (username, email, password_hash) VALUES (?, ?, ?)"
	_, err := s.db.Mgmt.ExecContext(r.Context(), insertSQL, req.Username, req.Email, string(hash))
	if err != nil {
		log.Printf("[%s] Error inserting admin: %v", s.Name(), err)
		http.Error(w, `{"error": "Failed to create master account"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[%s] Initialization complete. Master admin '%s' created.", s.Name(), req.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success", "message": "Master admin created."}`))
}