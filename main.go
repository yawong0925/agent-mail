// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent-mail/config"
	"agent-mail/database"
	"agent-mail/fetcher"
	"agent-mail/portal"
	"agent-mail/supervisor"
	"agent-mail/sysmgr"
	"agent-mail/wapi"
)

func main() {
	log.Println("Starting Agent Mail System...")

	// 1. Load Configuration
	cfg := config.LoadConfig()
	log.Printf("Web Portal: %s | Agent API: %s | SysMgr: %s", cfg.PortWebPortal, cfg.PortAgentAPI, cfg.PortSysMgr)

	// 2. Setup Context for Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Initialize Database Manager
	dbManager, err := database.NewManager(cfg)
	if err != nil {
		log.Fatalf("Fatal error initializing databases: %v", err)
	}
	defer dbManager.Close()

	// 4. Initialize Supervisor
	sup := supervisor.New()

	// 5. Register Modules
	sup.Register(wapi.New(cfg, dbManager))
	sup.Register(portal.New(cfg, dbManager))
	sup.Register(fetcher.NewIMAP(cfg, dbManager))
	sup.Register(sysmgr.New(cfg, dbManager))

	// 6. Listen for OS signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("\n[MAIN] Shutdown signal received. Initiating graceful shutdown...")
		cancel()
	}()

	// 7. Start Supervisor Watcher Loop (blocking call)
	sup.Run(ctx)
	
	// This will only print once the supervisor's Run() function unblocks
	log.Println("[MAIN] Reached complete shutdown. System exiting.")
	log.Println("Agent Mail System stopped cleanly.")
}