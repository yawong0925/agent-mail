// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-mail/config"
	"agent-mail/database"
	"agent-mail/fetcher"
	"agent-mail/portal"
	"agent-mail/supervisor"
	"agent-mail/sysmgr"
	"agent-mail/wapi"
)

func main() {
	log.Println("[MAIN] Starting Agent Mail System...")

	// 1. SETUP SIGNAL-AWARE ROOT CONTEXT
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 2. LOAD CONFIGURATION
	cfg := config.LoadConfig()

	// 3. INITIALIZE DATABASES
	dbManager, err := database.NewManager(cfg)
	if err != nil {
		log.Fatalf("[MAIN] Critical error initializing databases: %v\n", err)
	}
	defer dbManager.Close()
	log.Println("[DATABASE] SQLite databases initialized successfully.")

	// 4. INITIALIZE CENTRAL SUPERVISOR
	sup := supervisor.New()

	// 5. REGISTER MODULES (Using correct package constructors: .New())
	webPortal := portal.New(cfg, dbManager)
	sup.Register(webPortal)

	agentAPI := wapi.New(cfg, dbManager)
	sup.Register(agentAPI)

	sysMgr := sysmgr.New(cfg, dbManager)
	sup.Register(sysMgr)

	imapFetcher := fetcher.NewIMAP(cfg, dbManager)
	sup.Register(imapFetcher)

	log.Printf("[MAIN] Web Portal: %s | Agent API: %s | SysMgr: %s\n", cfg.PortWebPortal, cfg.PortAgentAPI, cfg.PortSysMgr)

	// 6. RUN THE SUPERVISOR
	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		sup.Run(ctx)
	}()

	// Wait here until the root context is cancelled by Ctrl+C
	<-ctx.Done()

	cancel() // restore default signal handling so a second Ctrl+C force-quits
	log.Println("[MAIN] Shutdown signal received. Initiating graceful shutdown...")

	// 7. EXPLICIT SHUTDOWN CASCADE
	select {
	case <-runDone:
	case <-time.After(10 * time.Second):
		log.Println("[MAIN] Timed out waiting for supervisor; forcing exit.")
	}

	log.Println("[MAIN] Reached complete shutdown. System exiting.")
}