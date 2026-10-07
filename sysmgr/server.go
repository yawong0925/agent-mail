// sysmgr/server.go
package sysmgr

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"agent-mail/config"
	"agent-mail/database"

	// We import the portal package to reuse its excellent MemorySessionStore
	"agent-mail/portal"
	"agent-mail/supervisor"
)

// Server implements supervisor.Module for the master System Management dashboard.
// It is completely isolated on port 8088.
type Server struct {
	cfg          *config.AppConfig
	db           *database.Manager
	server       *http.Server
	sessionStore *portal.MemorySessionStore // Isolated RAM map just for Admins
}

func New(cfg *config.AppConfig, db *database.Manager) *Server {
	return &Server{
		cfg:          cfg,
		db:           db,
		// This creates a brand new, empty RAM store separate from the port 8080 portal
		sessionStore: portal.NewMemorySessionStore(),
	}
}

func (s *Server) Name() string {
	return "SysMgr"
}

func (s *Server) Start(ctx context.Context, errChan chan<- supervisor.ModuleError) {
	// Start RAM garbage collection for admin sessions
	go s.sessionStore.StartGarbageCollector(ctx)

	mux := http.NewServeMux()

	// --- SETUP ROUTES (Public, but self-locking) ---
	mux.HandleFunc("GET /api/setup/status", s.handleSetupStatus)
	mux.HandleFunc("POST /api/setup/admin", s.handleSetupAdmin)
	
	// --- AUTH ROUTES ---
	mux.HandleFunc("POST /api/auth/login", s.handleAdminLogin) 

	// --- PROTECTED ADMIN ROUTES ---
	mux.HandleFunc("GET /api/system/telemetry", s.RequireAuth(s.handleDashboardStats))

	s.server = &http.Server{
		Addr:    s.cfg.PortSysMgr,
		// Wrap the router in CORS middleware so the Vue dev server can connect to it
		Handler: s.enableCORS(mux),
	}

	go func() {
		log.Printf("[%s] Master Dashboard listening on %s\n", s.Name(), s.cfg.PortSysMgr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- supervisor.ModuleError{Name: s.Name(), Err: err}
		}
	}()

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

func (s *Server) Status() string {
	return "Running"
}

func (s *Server) GetStats() map[string]interface{} {
	return map[string]interface{}{"port": s.cfg.PortSysMgr}
}