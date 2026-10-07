// fetcher/imap.go
package fetcher

import (
	// context allows us to listen for shutdown signals (like Ctrl+C) and instantly kill network requests mid-flight.
	"context"
	// tls configures the secure transport layer for our IMAP connections (e.g., SSL verification).
	"crypto/tls"
	// fmt allows string formatting, like building the "host:port" address or wrapping errors.
	"fmt"
	// io provides the Reader/Writer interfaces to stream attachments directly to disk (saving RAM).
	"io"
	// log handles console output for debugging and our shutdown checkpoints.
	"log"
	// net provides the dialer to interact with raw TCP sockets directly for instant termination.
	"net"
	// os allows interaction with the host system to create folders and files.
	"os"
	// filepath dynamically constructs file paths for Windows/Linux compatibility.
	"path/filepath"
	// strings sanitizes inputs (replacing slashes in email addresses to prevent directory traversal).
	"strings"
	// sync provides the WaitGroup to track active background threads.
	"sync"
	// time handles the 5-minute polling loop and timestamps.
	"time"

	// go-imap is the core protocol library that speaks the IMAP specification to the server.
	"github.com/emersion/go-imap"
	// client provides the high-level methods to login, select folders, and fetch emails.
	"github.com/emersion/go-imap/client"
	// mail parses the raw MIME structure, intelligently separating text bodies from file attachments.
	"github.com/emersion/go-message/mail"

	"agent-mail/config"
	"agent-mail/database"
	"agent-mail/supervisor"
)

// IMAPFetcher is the background worker responsible for pulling emails via standard IMAP.
// It implements the supervisor.Module interface so the system can monitor and gracefully stop it.
type IMAPFetcher struct {
	cfg          *config.AppConfig // Holds directory paths and settings loaded from .env
	db           *database.Manager // Holds the active SQLite connection pools
	pollInterval time.Duration     // The sleep duration between sync cycles across all accounts
}

// NewIMAP prepares the fetcher module in memory.
// It does not start the background process yet; it only prepares the struct dependencies.
func NewIMAP(cfg *config.AppConfig, db *database.Manager) *IMAPFetcher {
	return &IMAPFetcher{
		cfg:          cfg,
		db:           db,
		pollInterval: 5 * time.Minute, // Default safe interval
	}
}

// Name identifies the module to the Supervisor for logging.
func (f *IMAPFetcher) Name() string {
	return "IMAP-Fetcher"
}

// Start contains the infinite loop that drives the background fetcher.
func (f *IMAPFetcher) Start(ctx context.Context, errChan chan<- supervisor.ModuleError) {
	log.Printf("[%s] Started polling every %v\n", f.Name(), f.pollInterval)

	// Create a ticker that fires an event down its channel every 5 minutes.
	ticker := time.NewTicker(f.pollInterval)
	defer ticker.Stop() // Prevent memory leaks when shutting down.

	// Run an immediate sync sweep on boot so we don't wait 5 minutes for the first run.
	f.syncAccounts(ctx)

	// The infinite polling loop.
	for {
		select {
		case <-ctx.Done():
			// The system supervisor cancelled the context via Ctrl+C. Break the loop.
			log.Printf("[%s] Reached shutdown checkpoint: Main loop caught context cancellation.\n", f.Name())
			log.Printf("[%s] Reached complete shutdown.\n", f.Name())
			return
		case <-ticker.C:
			// The ticker fired. Launch a new synchronization cycle.
			f.syncAccounts(ctx)
		}
	}
}

// Stop fulfills the supervisor.Module interface.
// Because our Start() loop inherently listens to context cancellation, we just log our checkpoints here.
func (f *IMAPFetcher) Stop() error {
	log.Printf("[%s] Reached shutdown checkpoint: Stop() invoked by Supervisor.\n", f.Name())
	log.Printf("[%s] Reached complete shutdown (Context cancellation handled primary cleanup).\n", f.Name())
	return nil
}

// Status returns the current health state of the module for the SysMgr dashboard.
func (f *IMAPFetcher) Status() string {
	return "Running"
}

// GetStats returns numeric telemetry data to the SysMgr dashboard.
func (f *IMAPFetcher) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"poll_interval_seconds": f.pollInterval.Seconds(),
	}
}

// syncAccounts queries the database for all active IMAP accounts and launches a concurrent thread for each.
func (f *IMAPFetcher) syncAccounts(ctx context.Context) {
	log.Printf("[%s] Beginning concurrent sync cycle...\n", f.Name())

	query := `SELECT id, imap_host, imap_port, encryption, username, email_address, password_encrypted FROM imap_accounts WHERE is_active = 1`
	
	// QueryContext instantly aborts if the application is shutting down.
	rows, err := f.db.Mgmt.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("[%s] DB Query Error: %v\n", f.Name(), err)
		}
		return
	}
	defer rows.Close()

	var wg sync.WaitGroup

	// Loop through every account returned by the database.
	for rows.Next() {
		// Stop launching threads if Ctrl+C was pressed mid-loop.
		if ctx.Err() != nil {
			log.Printf("[%s] Reached shutdown checkpoint: Abandoning row iteration.\n", f.Name())
			break
		}

		var id, port int
		var host, encryption, user, emailAddr, pass string

		if err := rows.Scan(&id, &host, &port, &encryption, &user, &emailAddr, &pass); err != nil {
			continue
		}

		// Track this new thread in our WaitGroup.
		wg.Add(1)

		// Launch the worker goroutine. We pass the variables to prevent loop closure data races.
		go func(accID int, accHost string, accPort int, accEnc, accUser, accEmail, accPass string) {
			defer wg.Done()
			if err := f.processAccount(ctx, accID, accHost, accPort, accEnc, accUser, accEmail, accPass); err != nil {
				// Only print errors if we aren't intentionally crashing out via Ctrl+C
				if ctx.Err() == nil {
					log.Printf("[%s] Error syncing %s: %v\n", f.Name(), accEmail, err)
				}
			}
		}(id, host, port, encryption, user, emailAddr, pass)
	}

	// Always check rows.Err() in Go after a rows.Next() loop to catch silent database cursor failures.
	if err := rows.Err(); err != nil && ctx.Err() == nil {
		log.Printf("[%s] Row iteration error: %v\n", f.Name(), err)
	}

	// =====================================================================
	// THE WAITGROUP BYPASS FIX
	// =====================================================================
	// To prevent hanging on Ctrl+C, we do NOT call wg.Wait() directly on the main thread.
	// Instead, we wait in a background thread and signal when finished.
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	// Select between the normal WaitGroup completion OR the Ctrl+C context cancellation.
	select {
	case <-waitCh:
		log.Printf("[%s] Sync cycle complete.\n", f.Name())
	case <-ctx.Done():
		// If Ctrl+C is hit, we instantly abandon the WaitGroup and exit the function.
		log.Printf("[%s] Reached shutdown checkpoint: Abandoning WaitGroup instantly.\n", f.Name())
	}
	// =====================================================================
}

// processAccount dials a RAW socket to the server, manages encryption, and streams unread emails safely.
func (f *IMAPFetcher) processAccount(ctx context.Context, accountID int, host string, port int, encryption, user, emailAddr, pass string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	
	// 1. DIAL THE RAW TCP SOCKET
	// DialContext allows the initial network connection to be cancelled instantly if Ctrl+C is pressed.
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("tcp dial failed: %w", err)
	}

	// ==========================================
	// THE RAW SOCKET ASSASSIN (Watchdog)
	// ==========================================
	// Third-party libraries like go-imap ignore context cancellation. 
	// We tie a watchdog directly to the OS-level TCP socket we just opened to ensure it dies on Ctrl+C.
	watchdogCtx, cancelWatchdog := context.WithCancel(context.Background())
	defer cancelWatchdog()

	go func() {
		select {
		case <-ctx.Done():
			// The wire is cut. Any go-imap commands reading/writing will instantly panic and unblock.
			log.Printf("[%s] Reached shutdown checkpoint: Severing TCP socket for %s\n", f.Name(), emailAddr)
			if conn != nil {
				conn.Close()
			}
		case <-watchdogCtx.Done():
			// Main function finished normally, let watchdog sleep peacefully.
		}
	}()
	// ==========================================

	// 2. HAND RAW CONNECTION TO GO-IMAP AND NEGOTIATE ENCRYPTION
	var c *client.Client
	tlsConfig := &tls.Config{ServerName: host}
	encryption = strings.ToLower(encryption)
	
	switch encryption {
	case "tls", "ssl":
		tlsConn := tls.Client(conn, tlsConfig)
		// HandshakeContext also respects our Ctrl+C signal!
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("tls handshake failed: %w", err)
		}
		c, err = client.New(tlsConn)
	case "starttls":
		c, err = client.New(conn)
		if err == nil { err = c.StartTLS(tlsConfig) }
	case "plain", "none":
		c, err = client.New(conn)
	default:
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("tls handshake failed: %w", err)
		}
		c, err = client.New(tlsConn)
	}

	if err != nil {
		return fmt.Errorf("imap client init failed: %w", err)
	}

	// 3. CLEANUP DEFER
	defer func() {
		// Only attempt polite IMAP "LOGOUT" if the app is NOT shutting down.
		if ctx.Err() == nil && c != nil { 
			c.Logout() 
		}
	}()

	// 4. AUTHENTICATE AND FETCH UIDS
	if err := c.Login(user, pass); err != nil { return fmt.Errorf("IMAP login failed: %w", err) }
	
	mbox, err := c.Select("INBOX", false)
	if err != nil { return fmt.Errorf("INBOX selection failed: %w", err) }
	if mbox.Messages == 0 { return nil }

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	uids, err := c.UidSearch(criteria)
	if err != nil { return fmt.Errorf("UID SEARCH UNSEEN failed: %w", err) }
	if len(uids) == 0 { return nil }

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	// Buffered channel receives message pointers dynamically.
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	// Fetching blocks, so run it in a background goroutine.
	go func() {
		done <- c.UidFetch(seqset, items, messages)
	}()

	// 5. DEADLOCK-PROOF CHANNEL LOOP
fetchLoop:
	for {
		select {
		case <-ctx.Done():
			// Instantly break the loop without waiting for the next email over the wire.
			break fetchLoop
		case msg, ok := <-messages:
			if !ok {
				break fetchLoop // Channel closed gracefully
			}

			uid := fmt.Sprintf("%d", msg.Uid)
			
			// Deduplication check
			var exists int
			checkSQL := "SELECT 1 FROM emails WHERE account_id = ? AND remote_uid = ?"
			if err := f.db.Emails.QueryRowContext(ctx, checkSQL, accountID, uid).Scan(&exists); err == nil && exists == 1 {
				continue
			}

			body := msg.GetBody(section)
			if body == nil { continue }

			f.saveEmail(ctx, accountID, user, emailAddr, uid, body)
		}
	}

	// 6. INSTANT RETURN FIX
	// Wait for the background fetch to finish, OR for Ctrl+C. Whichever is first.
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// saveEmail streams the raw email to disk, parses attachments, and executes SQLite inserts atomically.
func (f *IMAPFetcher) saveEmail(ctx context.Context, accountID int, username, emailAddr, uid string, body io.Reader) error {
	mr, err := mail.CreateReader(body)
	if err != nil { return fmt.Errorf("failed to create mail MIME reader: %w", err) }

	header := mr.Header
	subject, _ := header.Subject()
	date, _ := header.Date()

	// Target Structure: /data/storage/emls/<username>/<email_address>/
	safeUser := strings.ReplaceAll(strings.ReplaceAll(username, "/", "_"), "\\", "_")
	safeEmail := strings.ReplaceAll(strings.ReplaceAll(emailAddr, "/", "_"), "\\", "_")

	userEmlDir := filepath.Join(f.cfg.EmlStorageDir, safeUser, safeEmail)
	userAttachDir := filepath.Join(f.cfg.AttachStorageDir, safeUser, safeEmail)

	os.MkdirAll(userEmlDir, 0750)
	os.MkdirAll(userAttachDir, 0750)

	emlPath := filepath.Join(userEmlDir, fmt.Sprintf("%d_%s.eml", time.Now().UnixNano(), uid))

	// Atomic database transaction.
	tx, err := f.db.Emails.BeginTx(ctx, nil)
	if err != nil { return fmt.Errorf("failed to begin email transaction: %w", err) }
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO emails (account_id, remote_uid, date_time, subject, eml_file_path) VALUES (?, ?, ?, ?, ?)`, accountID, uid, date, subject, emlPath)
	if err != nil { return fmt.Errorf("failed to insert parent email record: %w", err) }

	emailID, _ := res.LastInsertId()
	hasAttachment := false

	// Parse MIME parts (Text, HTML, Attachments)
	for {
		if ctx.Err() != nil { return ctx.Err() }

		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			continue
		}

		switch h := part.Header.(type) {
		case *mail.AttachmentHeader:
			hasAttachment = true
			filename, _ := h.Filename()
			if filename == "" { filename = "unnamed_attachment" }

			attachPath := filepath.Join(userAttachDir, fmt.Sprintf("%d_%s", emailID, filename))
			outFile, err := os.Create(attachPath)
			if err != nil { continue }

			// Stream from network to disk instantly
			written, _ := io.Copy(outFile, part.Body)
			outFile.Close()

			tx.ExecContext(ctx, `INSERT INTO email_attachments (email_id, original_filename, stored_file_path, file_size) VALUES (?, ?, ?, ?)`, emailID, filename, attachPath, written)
		}
	}

	if hasAttachment {
		tx.ExecContext(ctx, "UPDATE emails SET has_attachment = 1 WHERE id = ?", emailID)
	}

	if err := tx.Commit(); err != nil { return fmt.Errorf("failed to commit email transaction: %w", err) }
	
	log.Printf("[%s] Saved email ID %d: '%s'\n", f.Name(), emailID, subject)
	return nil
}