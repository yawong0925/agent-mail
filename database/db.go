// database/db.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// Import the SQLite driver so database/sql can use it implicitly
	_ "github.com/mattn/go-sqlite3"

	"agent-mail/config"
)

// Manager holds the active SQLite connection pools.
type Manager struct {
	Emails *sql.DB
	Auth   *sql.DB
	Mgmt   *sql.DB
}

// NewManager ensures storage directories exist, connects to SQLite files, and executes schemas.
func NewManager(cfg *config.AppConfig) (*Manager, error) {
	// 1. Prepare directory paths based on the .env config
	dirs := []string{
		filepath.Dir(cfg.DBPathEmails),
		filepath.Dir(cfg.DBPathAuth),
		filepath.Dir(cfg.DBPathMgmt),
		cfg.EmlStorageDir,
		cfg.AttachStorageDir,
		cfg.LogStorageDir,
	}

	// 2. Ensure directories exist on the host system
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 3. Connect to the SQLite databases
	// We use WAL (Write-Ahead Logging) to allow multiple concurrent readers/writers without locking the whole file.
	// We use a busy_timeout of 5 seconds to queue queries rather than failing instantly if a lock occurs.
	dbEmails, err := sql.Open("sqlite3", cfg.DBPathEmails+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open emails db: %w", err)
	}

	dbAuth, err := sql.Open("sqlite3", cfg.DBPathAuth+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open auth db: %w", err)
	}

	dbMgmt, err := sql.Open("sqlite3", cfg.DBPathMgmt+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open mgmt db: %w", err)
	}

	m := &Manager{
		Emails: dbEmails,
		Auth:   dbAuth,
		Mgmt:   dbMgmt,
	}

	// 4. Build the tables using the schemas defined in schema.go
	if err := m.buildSchemas(); err != nil {
		return nil, err
	}

	log.Println("[DATABASE] SQLite databases initialized successfully.")
	return m, nil
}

// Close gracefully shuts down the database connection pools during system exit.
func (m *Manager) Close() {
	if m.Emails != nil { m.Emails.Close() }
	if m.Auth != nil { m.Auth.Close() }
	if m.Mgmt != nil { m.Mgmt.Close() }
	log.Println("[DATABASE] Connections closed.")
}

// buildSchemas executes the SQL table definitions stored in schema.go.
func (m *Manager) buildSchemas() error {
	// Execute schemaEmails against DB1
	if _, err := m.Emails.Exec(schemaEmails); err != nil {
		return fmt.Errorf("error building emails schema: %w", err)
	}

	// Execute schemaAuth against DB2
	if _, err := m.Auth.Exec(schemaAuth); err != nil {
		return fmt.Errorf("error building auth schema: %w", err)
	}

	// Execute schemaMgmt against DB3
	if _, err := m.Mgmt.Exec(schemaMgmt); err != nil {
		return fmt.Errorf("error building mgmt schema: %w", err)
	}

	return nil
}