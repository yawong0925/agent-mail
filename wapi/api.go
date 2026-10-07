// wapi/api.go
package wapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"agent-mail/config"
	"agent-mail/database"
	"agent-mail/supervisor"
)

// API holds the dependencies required for the HTTP server to run.
// It implements the supervisor.Module interface.
type API struct {
	cfg    *config.AppConfig
	db     *database.Manager
	server *http.Server
}

// New initializes a new instance of the WAPI module.
func New(cfg *config.AppConfig, db *database.Manager) *API {
	return &API{
		cfg: cfg,
		db:  db,
	}
}

// Name provides the identifier used by the Supervisor.
func (a *API) Name() string {
	return "Agent-API"
}

// Start boots up the HTTP server and blocks until the context is cancelled.
func (a *API) Start(ctx context.Context, errChan chan<- supervisor.ModuleError) {
	// Create a new HTTP multiplexer using standard Go 1.22+ routing syntax
	mux := http.NewServeMux()
	
	// Register the endpoints
	mux.HandleFunc("GET /api/agent/emails", a.handleListEmails)
	mux.HandleFunc("GET /api/agent/emails/{id}/download", a.handleDownloadEmail)

	// Configure the HTTP server
	a.server = &http.Server{
		Addr:    a.cfg.PortAgentAPI,
		Handler: mux,
	}

	// Spin up the server in a separate goroutine so we don't block the supervisor context listener
	go func() {
		log.Printf("[%s] Server listening on %s\n", a.Name(), a.cfg.PortAgentAPI)
		
		// ListenAndServe always returns an error when it stops.
		// We ignore http.ErrServerClosed because that means we shut it down intentionally via a.Stop().
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- supervisor.ModuleError{Name: a.Name(), Err: err}
		}
	}()

	// Block here and listen for the OS shutdown signal propagated by the Supervisor
	<-ctx.Done()
	a.Stop()
}

// Stop initiates a graceful shutdown of the HTTP server.
func (a *API) Stop() error {
	if a.server != nil {
		log.Printf("[%s] Reached shutdown checkpoint: Draining HTTP connections...\n", a.Name())
		
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := a.server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("[%s] Shutdown error: %v\n", a.Name(), err)
			return err
		}

		log.Printf("[%s] Reached complete shutdown.\n", a.Name())
		return nil
	}

	log.Printf("[%s] Reached complete shutdown (Server was not running).\n", a.Name())
	return nil
}

// Status returns a simple string representing current health.
func (a *API) Status() string {
	return "Running"
}

// GetStats returns telemetry data for the System Manager panel.
func (a *API) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"port":   a.cfg.PortAgentAPI,
		"status": "active",
	}
}

// --- Route Handlers ---

// handleListEmails returns a JSON list of available emails.
func (a *API) handleListEmails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: Wire this up to a.db.Emails.Query()
	w.Write([]byte(`{"status":"success", "message":"Email list endpoint reached."}`))
}

// handleDownloadEmail prepares an email and its attachments for the AI Agent.
func (a *API) handleDownloadEmail(w http.ResponseWriter, r *http.Request) {
	// Extract the {id} path variable provided by Go 1.22's router
	id := r.PathValue("id")
	
	w.Header().Set("Content-Type", "application/json")
	// TODO: Implement XZ compression logic here
	w.Write([]byte(`{"status":"success", "message":"Download endpoint reached for ID: ` + id + `"}`))
}