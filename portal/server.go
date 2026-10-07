// portal/server.go
package portal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"agent-mail/config"
	"agent-mail/database"
	"agent-mail/supervisor"
)

// Server implements the supervisor.Module interface.
// It manages the HTTP API utilized by the Vue 3 frontend administration dashboard.
type Server struct {
	cfg          *config.AppConfig
	db           *database.Manager
	server       *http.Server
	sessionStore *MemorySessionStore // Points to the thread-safe RAM map we created
}

// New constructs the Web Portal module and initializes its memory stores.
func New(cfg *config.AppConfig, db *database.Manager) *Server {
	return &Server{
		cfg:          cfg,
		db:           db,
		sessionStore: NewMemorySessionStore(),
	}
}

// Name returns the identifier used by the central Supervisor for logging and crash recovery.
func (s *Server) Name() string {
	return "Web-Portal"
}

// Start launches the HTTP server and background maintenance tasks.
func (s *Server) Start(ctx context.Context, errChan chan<- supervisor.ModuleError) {
	// 1. Boot the background RAM garbage collector.
	// This sweeps memory every 10 minutes to delete expired session tokens.
	go s.sessionStore.StartGarbageCollector(ctx)

	// 2. Initialize the standard Go 1.22+ HTTP multiplexer.
	mux := http.NewServeMux()

	// --- PUBLIC ROUTES (No Authentication Required) ---
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/verify", s.handleVerify2FA)

	// --- PROTECTED ROUTES (Authentication Required) ---
	// Wrapping a handler in s.RequireAuth forces the client to provide a valid session token.
	mux.HandleFunc("POST /api/accounts/imap", s.RequireAuth(s.handleAddIMAPAccount))

	// Configure the HTTP server bound to the configuration port (default :8080).
	s.server = &http.Server{
		Addr:    s.cfg.PortWebPortal,
		Handler: mux,
	}

	// 3. Start listening in a background goroutine so we don't block the supervisor.
	go func() {
		log.Printf("[%s] Server listening on %s\n", s.Name(), s.cfg.PortWebPortal)
		
		// If ListenAndServe returns an error other than ErrServerClosed, a fatal crash occurred.
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- supervisor.ModuleError{Name: s.Name(), Err: err}
		}
	}()

	// 4. Block execution here until the application begins a graceful shutdown.
	<-ctx.Done()
	s.Stop()
}

func (s *Server) Stop() error {
	if s.server != nil {
		log.Printf("[%s] Reached shutdown checkpoint: Draining HTTP connections...\n", s.Name())
		
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := s.server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("[%s] Shutdown error: %v\n", s.Name(), err)
			return err
		}
		
		log.Printf("[%s] Reached complete shutdown.\n", s.Name())
		return nil
	}
	
	log.Printf("[%s] Reached complete shutdown (Server was not running).\n", s.Name())
	return nil
}

// Status provides a simple health check string.
func (s *Server) Status() string {
	return "Running"
}

// GetStats compiles basic telemetry for the portal.
func (s *Server) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"port": s.cfg.PortWebPortal,
	}
}

// --- Route Handlers ---

// handleIndex serves a simple availability check at the root URL.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "online", "service": "Agent Mail Web Portal"}`))
}

// AccountRequest defines the expected JSON payload for manually adding a new IMAP account.
type AccountRequest struct {
	EmailAddress string `json:"email_address"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	IMAPHost     string `json:"imap_host"`
	IMAPPort     int    `json:"imap_port"`
	Encryption   string `json:"encryption"`
}

// handleAddIMAPAccount parses the JSON payload and saves the new account directly to SQLite.
func (s *Server) handleAddIMAPAccount(w http.ResponseWriter, r *http.Request) {
	var req AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	// Validate required inputs.
	if req.EmailAddress == "" || req.IMAPHost == "" || req.Password == "" {
		http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		return
	}

	// Apply default values if omitted by the client.
	if req.Encryption == "" { req.Encryption = "tls" }
	if req.Username == "" { req.Username = req.EmailAddress }

	// Insert into DB3 (Mgmt) - The table was recently renamed to 'imap_accounts'.
	insertSQL := `
		INSERT INTO imap_accounts 
		(email_address, username, password_encrypted, imap_host, imap_port, encryption) 
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err := s.db.Mgmt.ExecContext(r.Context(), insertSQL, 
		req.EmailAddress, req.Username, req.Password, req.IMAPHost, req.IMAPPort, req.Encryption,
	)
	
	if err != nil {
		log.Printf("[%s] Database error adding account: %v", s.Name(), err)
		http.Error(w, `{"error": "Failed to save account"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success", "message": "IMAP Account added."}`))
}